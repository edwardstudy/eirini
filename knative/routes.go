package knative

import (
	"code.cloudfoundry.org/eirini"
	"code.cloudfoundry.org/eirini/route"
	servingclientset "github.com/knative/serving/pkg/client/clientset/versioned"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ServiceRouteLister struct {
	client    servingclientset.Interface
	namespace string
}

func NewServiceRouteLister(client servingclientset.Interface, namespace string) route.Lister {
	return &ServiceRouteLister{
		client:    client,
		namespace: namespace,
	}
}

func (r *ServiceRouteLister) ListRoutes() ([]*eirini.Routes, error) {
	services, err := r.client.Serving().Services(r.namespace).List(meta_v1.ListOptions{})
	if err != nil {
		return []*eirini.Routes{}, err
	}

	routes := []*eirini.Routes{}
	for _, s := range services.Items {
		if !isCFService(s.Name) {
			continue
		}

		registered, err := decodeRoutes(s.Annotations[eirini.RegisteredRoutes])
		if err != nil {
			return []*eirini.Routes{}, err
		}

		route := eirini.Routes{
			Routes: registered,
			Name:   s.Name,
		}

		routes = append(routes, &route)
	}

	return routes, nil
}
