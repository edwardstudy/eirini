package knative

import (
	"encoding/json"
	"fmt"
	"regexp"

	"code.cloudfoundry.org/eirini"
	"code.cloudfoundry.org/eirini/models/cf"
	"code.cloudfoundry.org/eirini/opi"
	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	servingclientset "github.com/knative/serving/pkg/client/clientset/versioned"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	knife "github.com/julz/knife/pkg/knative"
	types "github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
)

type serviceManager struct {
	client     servingclientset.Interface
	namespace  string
	routesChan chan<- []*eirini.Routes
}

func NewServiceManager(client servingclientset.Interface, namespace string, routesChan chan<- []*eirini.Routes) ServiceManager {
	return &serviceManager{
		client:     client,
		namespace:  namespace,
		routesChan: routesChan,
	}
}

func (m *serviceManager) services() types.ServiceInterface {
	return m.client.Serving().Services(m.namespace)
}

func (m *serviceManager) Exists(appName string) (bool, error) {
	selector := fmt.Sprintf("name=%s", appName)
	list, err := m.services().List(meta_v1.ListOptions{LabelSelector: selector})
	if err != nil {
		return false, err
	}

	return len(list.Items) > 0, nil
}

func (m *serviceManager) Create(lrp *opi.LRP) error {
	s, err := m.services().Create(toService(lrp))
	if err != nil {
		return err
	}

	registeredRoutes, err := decodeRoutes(s.Annotations[eirini.RegisteredRoutes])
	if err != nil {
		return err
	}

	routes := eirini.Routes{
		Routes: registeredRoutes,
		Name:   s.Name,
	}

	m.routesChan <- []*eirini.Routes{&routes}
	return nil
}

func (m *serviceManager) Update(lrp *opi.LRP) error {
	serviceName := eirini.GetInternalServiceName(lrp.Name)
	service, err := m.services().Get(serviceName, meta_v1.GetOptions{})
	if err != nil {
		return err
	}

	registeredRoutes, err := decodeRoutes(service.Annotations[eirini.RegisteredRoutes])
	if err != nil {
		return err
	}
	updatedRoutes, err := decodeRoutes(lrp.Metadata[cf.VcapAppUris])
	if err != nil {
		return err
	}

	service.Annotations[eirini.RegisteredRoutes] = lrp.Metadata[cf.VcapAppUris]
	for k, v := range lrp.Metadata {
		service.Annotations[k] = v
	}

	_, err = m.services().Update(service)
	if err != nil {
		return err
	}

	unregistered := getUnregisteredRoutes(registeredRoutes, updatedRoutes)
	routes := eirini.Routes{
		Routes:             updatedRoutes,
		UnregisteredRoutes: unregistered,
		Name:               serviceName,
	}

	m.routesChan <- []*eirini.Routes{&routes}
	return nil
}

func (m *serviceManager) Delete(appName string) error {
	serviceName := eirini.GetInternalServiceName(appName)
	service, err := m.services().Get(serviceName, meta_v1.GetOptions{})
	if err != nil {
		return err
	}

	existingRoutes, err := decodeRoutes(service.Annotations[eirini.RegisteredRoutes])
	if err != nil {
		return err
	}

	routes := eirini.Routes{
		UnregisteredRoutes: existingRoutes,
		Name:               serviceName,
	}

	m.routesChan <- []*eirini.Routes{&routes}

	backgroundPropagation := meta_v1.DeletePropagationBackground
	return m.services().Delete(serviceName, &meta_v1.DeleteOptions{PropagationPolicy: &backgroundPropagation})
}

func toService(lrp *opi.LRP) *v1alpha1.Service {
	service := knife.NewRunLatestService(eirini.GetInternalServiceName(lrp.Name),
		knife.WithRevisionTemplate(lrp.Image, lrp.Command, lrp.Env),
	)

	service.Labels = map[string]string{
		"name":   lrp.Name,
		"eirini": "eirini",
	}

	service.Annotations = map[string]string{
		eirini.RegisteredRoutes: lrp.Metadata[cf.VcapAppUris],
	}

	for k, v := range lrp.Metadata {
		service.Annotations[k] = v
	}

	return service
}

func getUnregisteredRoutes(existing, updated []string) []string {
	updatedMap := sliceToMap(updated)
	unregistered := []string{}
	for _, e := range existing {
		if _, ok := updatedMap[e]; !ok {
			unregistered = append(unregistered, e)
		}
	}

	return unregistered
}

func sliceToMap(slice []string) map[string]bool {
	result := make(map[string]bool, len(slice))
	for _, e := range slice {
		result[e] = true
	}
	return result
}

func decodeRoutes(s string) ([]string, error) {
	uris := []string{}
	err := json.Unmarshal([]byte(s), &uris)

	return uris, err
}

func isCFService(s string) bool {
	return matchRegex(s, "^cf-.*$")
}

func matchRegex(subject string, regex string) bool {
	r, err := regexp.Compile(regex)
	if err != nil {
		panic(err)
	}
	return r.MatchString(subject)

}
