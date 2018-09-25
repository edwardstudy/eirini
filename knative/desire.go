package knative

import (
	"code.cloudfoundry.org/eirini"
	"code.cloudfoundry.org/eirini/opi"
	servingclientset "github.com/knative/serving/pkg/client/clientset/versioned"
	"k8s.io/client-go/kubernetes"
)

type Desirer struct {
	ServiceManager ServiceManager
	ServiceGetter  ServiceGetter
}

//go:generate counterfeiter . ServiceGetter
type ServiceGetter interface {
	List() ([]*opi.LRP, error)
	Get(name string) (*opi.LRP, error)
}

//go:generate counterfeiter . ServiceManager
type ServiceManager interface {
	Create(lrp *opi.LRP) error
	Update(lrp *opi.LRP) error
	Delete(appName string) error
	Exists(appName string) (bool, error)
}

func NewDesirer(kubeNamespace string, clientset kubernetes.Interface, servingclient servingclientset.Interface, routesChan chan []*eirini.Routes) *Desirer {
	return &Desirer{
		ServiceManager: NewServiceManager(servingclient, kubeNamespace, routesChan),
		ServiceGetter:  NewReplicaSetGetter(clientset, servingclient, kubeNamespace),
	}
}

func (d *Desirer) Desire(lrp *opi.LRP) error {
	exists, err := d.ServiceManager.Exists(lrp.Name)
	if err != nil || exists {
		return err
	}

	return d.ServiceManager.Create(lrp)
}

func (d *Desirer) List() ([]*opi.LRP, error) {
	return d.ServiceGetter.List()
}

func (d *Desirer) Get(name string) (*opi.LRP, error) {
	return d.ServiceGetter.Get(name)
}

func (d *Desirer) Update(lrp *opi.LRP) error {
	return d.ServiceManager.Update(lrp)
}

func (d *Desirer) Stop(name string) error {
	return d.ServiceManager.Delete(name)
}
