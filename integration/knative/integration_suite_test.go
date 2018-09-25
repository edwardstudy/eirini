package integration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	servingclientset "github.com/knative/serving/pkg/client/clientset/versioned"
	types "github.com/knative/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/api/apps/v1beta2"
	"k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	coretypes "k8s.io/client-go/kubernetes/typed/apps/v1beta2"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	timeout time.Duration = 60 * time.Second
)

var (
	namespace     string
	clientset     kubernetes.Interface
	servingClient servingclientset.Interface
)

var _ = BeforeSuite(func() {
	config, err := clientcmd.BuildConfigFromFlags("",
		filepath.Join(os.Getenv("HOME"), ".kube", "config"),
	)
	Expect(err).ToNot(HaveOccurred())

	clientset, err = kubernetes.NewForConfig(config)
	Expect(err).ToNot(HaveOccurred())
	clientset, err = kubernetes.NewForConfig(config)
	Expect(err).ToNot(HaveOccurred())

	servingClient, err = servingclientset.NewForConfig(config)
	Expect(err).ToNot(HaveOccurred())

	namespace = "opi-integration"
	if !namespaceExists() {
		createNamespace()
	}
})

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

func namespaceExists() bool {
	_, err := clientset.CoreV1().Namespaces().Get(namespace, meta.GetOptions{})
	return err == nil
}

func createNamespace() {
	namespaceSpec := &v1.Namespace{
		ObjectMeta: meta.ObjectMeta{Name: namespace},
	}

	if _, err := clientset.CoreV1().Namespaces().Create(namespaceSpec); err != nil {
		panic(err)
	}
}

func replicaSets() coretypes.ReplicaSetInterface {
	return clientset.AppsV1beta2().ReplicaSets(namespace)
}

func services() types.ServiceInterface {
	return servingClient.Serving().Services(namespace)
}

func cleanupService(appName string) {
	if _, err := services().Get(appName, meta.GetOptions{}); err == nil {
		backgroundPropagation := meta.DeletePropagationBackground
		err = services().Delete(appName, &meta.DeleteOptions{PropagationPolicy: &backgroundPropagation})
		Expect(err).ToNot(HaveOccurred())
	}
}

func listAllServices() []v1alpha1.Service {
	list, err := services().List(meta.ListOptions{})
	Expect(err).NotTo(HaveOccurred())
	return list.Items
}

func listService(appName string) []v1alpha1.Service {
	labelSelector := fmt.Sprintf("name=%s", appName)
	list, err := services().List(meta.ListOptions{LabelSelector: labelSelector})
	Expect(err).NotTo(HaveOccurred())
	return list.Items
}

func listAllReplicaSets() []v1beta2.ReplicaSet {
	list, err := clientset.AppsV1beta2().ReplicaSets(namespace).List(meta.ListOptions{})
	Expect(err).NotTo(HaveOccurred())
	return list.Items
}

func listReplicaSets(appName string) []v1beta2.ReplicaSet {
	labelSelector := fmt.Sprintf("name=%s", appName)
	list, err := clientset.AppsV1beta2().ReplicaSets(namespace).List(meta.ListOptions{LabelSelector: labelSelector})
	Expect(err).NotTo(HaveOccurred())
	return list.Items
}

func listPods(appName string) []v1.Pod {
	labelSelector := fmt.Sprintf("name=%s", appName)
	pods, err := clientset.CoreV1().Pods(namespace).List(meta.ListOptions{LabelSelector: labelSelector})
	Expect(err).NotTo(HaveOccurred())
	return pods.Items
}

func getPodNames(appName string) []string {
	names := []string{}
	for _, p := range listPods(appName) {
		names = append(names, p.Name)
	}

	return names
}
