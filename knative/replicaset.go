package knative

import (
	"code.cloudfoundry.org/eirini"
	"code.cloudfoundry.org/eirini/opi"
	"github.com/knative/serving/pkg/apis/serving"
	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	servingclientset "github.com/knative/serving/pkg/client/clientset/versioned"
	v1alpha1types "github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
	"github.com/pkg/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	types "k8s.io/client-go/kubernetes/typed/apps/v1beta2"
)

type replicaSetGetter struct {
	client        kubernetes.Interface
	servingClient servingclientset.Interface
	namespace     string
}

func NewReplicaSetGetter(client kubernetes.Interface, servingClient servingclientset.Interface, namespace string) ServiceGetter {
	return &replicaSetGetter{
		client:        client,
		servingClient: servingClient,
		namespace:     namespace,
	}
}

// FIXME - This is really inefficient
func (m *replicaSetGetter) List() ([]*opi.LRP, error) {
	services, err := m.services().List(meta.ListOptions{})
	if err != nil {
		return nil, err
	}

	lrps := make([]*opi.LRP, 0, len(services.Items))

	for _, s := range services.Items {
		lrp, err := m.serviceToLRP(&s)
		if err != nil {
			return nil, err
		}
		lrps = append(lrps, lrp)
	}

	return lrps, nil
}

func (m *replicaSetGetter) Get(appName string) (*opi.LRP, error) {
	service, err := m.services().Get(eirini.GetInternalServiceName(appName), meta.GetOptions{})
	if err != nil {
		return nil, err
	}

	return m.serviceToLRP(service)
}

func (m *replicaSetGetter) replicaSets() types.ReplicaSetInterface {
	return m.client.AppsV1beta2().ReplicaSets(m.namespace)
}

func (m *replicaSetGetter) services() v1alpha1types.ServiceInterface {
	return m.servingClient.Serving().Services(m.namespace)
}

func (m *replicaSetGetter) serviceToLRP(s *v1alpha1.Service) (*opi.LRP, error) {
	cenv := s.Spec.RunLatest.Configuration.RevisionTemplate.Spec.Container.Env
	env := make(map[string]string)
	filteredVars := map[string]bool{"POD_NAME": true}

	for _, v := range cenv {
		if !filteredVars[v.Name] {
			env[v.Name] = v.Value
		}
	}

	lrp := &opi.LRP{
		Name:    s.Labels["name"],
		Image:   s.Spec.RunLatest.Configuration.RevisionTemplate.Spec.Container.Image,
		Command: s.Spec.RunLatest.Configuration.RevisionTemplate.Spec.Container.Args,
		Env:     env,
		// TODO - Find out how to get the HealthCheck from Service (it should be easy)
		// Health: opi.Healtcheck{
		// 	Type:      "",
		// 	Port:      0,
		// 	Endpoint:  "",
		// 	TimeoutMs: 0,
		// },
		RunningInstances: 0,
		Metadata:         s.Annotations,
	}

	lastRevisionName := s.Status.LatestCreatedRevisionName

	if s.Status.LatestReadyRevisionName != "" {
		lastRevisionName = s.Status.LatestReadyRevisionName
	}

	if lastRevisionName == "" {
		return lrp, nil
	}
	list, err := m.replicaSets().List(meta.ListOptions{
		LabelSelector: labels.Set(map[string]string{
			serving.RevisionLabelKey: lastRevisionName,
		}).String(),
	})
	if err != nil {
		return nil, err
	}

	switch len(list.Items) {
	case 0:
		// TODO - What to do in this case depends how we interact with the rest of the system. (We could also define an error type)
		// Option 1
		return lrp, nil
		// Option 2
		// return nil, errors.Errorf("ReplicaSets not found for configuration %s", s.Name)
	case 1:
		lrp.RunningInstances = int(list.Items[0].Status.ReadyReplicas)
		lrp.TargetInstances = int(list.Items[0].Status.Replicas)
		return lrp, nil
	default:
		return nil, errors.Errorf("Too many ReplicaSets found for %s: %d", s.Name, len(list.Items))
	}
}
