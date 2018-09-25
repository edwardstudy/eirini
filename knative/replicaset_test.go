package knative_test

import (
	. "code.cloudfoundry.org/eirini/knative"
	"code.cloudfoundry.org/eirini/models/cf"
	"code.cloudfoundry.org/eirini/opi"
	"github.com/knative/serving/pkg/apis/serving"
	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	servingclientset "github.com/knative/serving/pkg/client/clientset/versioned"
	servingfake "github.com/knative/serving/pkg/client/clientset/versioned/fake"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/api/apps/v1beta2"
	"k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

var _ = Describe("ReplicaSet", func() {
	var (
		err           error
		client        kubernetes.Interface
		servingClient servingclientset.Interface
		getter        ServiceGetter
	)

	BeforeEach(func() {
		client = fake.NewSimpleClientset()
		servingClient = servingfake.NewSimpleClientset()
		getter = NewReplicaSetGetter(client, servingClient, namespace)
	})

	Context("When getting an app", func() {
		var name string
		var lrp *opi.LRP

		BeforeEach(func() {
			s := toService(createLRP("odin", "1234.5", "my.example.route"), namespace)
			_, createErr := servingClient.Serving().Services(namespace).Create(s)
			Expect(createErr).ToNot(HaveOccurred())
			name = "odin"
		})

		JustBeforeEach(func() {
			lrp, err = getter.Get(name)
		})

		It("return the expected LRP", func() {
			Expect(err).NotTo(HaveOccurred())
			lrp, err = getter.Get(name)
			Expect(err).ToNot(HaveOccurred())
			Expect(lrp).To(Equal(
				&opi.LRP{
					Name: "odin",
					Command: []string{
						"/bin/sh",
						"-c",
						"while true; do echo hello; sleep 10;done",
					},
					Env:              map[string]string{"env": "env-value"},
					TargetInstances:  0,
					RunningInstances: 0,
					Image:            "busybox",
					Metadata: map[string]string{
						cf.ProcessGUID: "odin",
						cf.LastUpdated: "1234.5",
						cf.VcapAppUris: "my.example.route",
						"routes":       "my.example.route",
					},
				},
			))
		})

		Context("When an LRP doesn't have a ready revision", func() {
			BeforeEach(func() {
				l := createLRP("zeus", "irrelevant", "irrelevant")
				service := toService(l, namespace)
				service.Status = v1alpha1.ServiceStatus{LatestCreatedRevisionName: "zeus-v1"}
				_, createErr := servingClient.Serving().Services(namespace).Create(service)
				Expect(createErr).ToNot(HaveOccurred())
				_, createErr = client.AppsV1beta2().ReplicaSets(namespace).Create(toReplicaSet(l, "zeus-v1", 2, 1, 0))
				Expect(createErr).ToNot(HaveOccurred())
				name = "zeus"
			})

			It("returns target replicas from replicaset", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(lrp.RunningInstances).To(Equal(0))
				Expect(lrp.TargetInstances).To(Equal(2))
			})
		})

		Context("When an LRP has a revision ready", func() {
			var want *opi.LRP

			BeforeEach(func() {
				want = createLRP("zeus", "irrelevant", "irrelevant")
				service := toService(want, namespace)
				service.Status = v1alpha1.ServiceStatus{LatestReadyRevisionName: "zeus-v1"}
				_, createErr := servingClient.Serving().Services(namespace).Create(service)
				Expect(createErr).ToNot(HaveOccurred())
				name = "zeus"
			})

			Context("when a replicaset has been created", func() {
				BeforeEach(func() {
					_, createErr := client.AppsV1beta2().ReplicaSets(namespace).Create(toReplicaSet(want, "zeus-v1", 2, 2, 1))
					Expect(createErr).ToNot(HaveOccurred())
				})

				It("returns target and running instances from replicaset", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(lrp.RunningInstances).To(Equal(1))
					Expect(lrp.TargetInstances).To(Equal(2))
				})
			})

			Context("when the replicaset is not found", func() {
				It("returns the target and running instances as zero", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(lrp.RunningInstances).To(Equal(0))
					Expect(lrp.TargetInstances).To(Equal(0))
				})
			})
		})

		Context("When the app has several revisions", func() {
			BeforeEach(func() {
				l := createLRP("zeus", "irrelevant", "irrelevant")
				service := toService(l, namespace)
				service.Status = v1alpha1.ServiceStatus{LatestReadyRevisionName: "zeus-v2"}
				_, createErr := servingClient.Serving().Services(namespace).Create(service)
				Expect(createErr).ToNot(HaveOccurred())
				_, createErr = client.AppsV1beta2().ReplicaSets(namespace).Create(toReplicaSet(l, "zeus-v1", 2, 0, 0))
				Expect(createErr).ToNot(HaveOccurred())
				_, createErr = client.AppsV1beta2().ReplicaSets(namespace).Create(toReplicaSet(l, "zeus-v2", 2, 2, 1))
				Expect(createErr).ToNot(HaveOccurred())
				name = "zeus"
			})

			It("returns the number of instances of the latest ready", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(lrp.RunningInstances).To(Equal(1))
				Expect(lrp.TargetInstances).To(Equal(2))
			})
		})

		Context("when the app does not exist", func() {
			JustBeforeEach(func() {
				lrp, err = getter.Get("non-existent")
			})

			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Context("List", func() {
		var lrps []*opi.LRP

		Context("when some services exist", func() {
			BeforeEach(func() {
				lrps = []*opi.LRP{
					createLRP("odin", "1234.5", "my.example.route"),
					createLRP("thor", "4567.8", "my.example.route"),
				}

				for _, l := range lrps {
					s := toService(l, namespace)
					_, createErr := servingClient.Serving().Services(namespace).Create(s)
					Expect(createErr).ToNot(HaveOccurred())
				}
			})

			It("translates to opi.LRPs", func() {
				actualLRPs, err := getter.List()
				Expect(err).ToNot(HaveOccurred())

				Expect(actualLRPs).To(ConsistOf(
					&opi.LRP{
						Name: "odin",
						Command: []string{
							"/bin/sh",
							"-c",
							"while true; do echo hello; sleep 10;done",
						},
						Env:   map[string]string{"env": "env-value"},
						Image: "busybox",
						Metadata: map[string]string{
							cf.ProcessGUID: "odin",
							cf.LastUpdated: "1234.5",
							cf.VcapAppUris: "my.example.route",
							"routes":       "my.example.route",
						},
					},
					&opi.LRP{
						Name: "thor",
						Command: []string{
							"/bin/sh",
							"-c",
							"while true; do echo hello; sleep 10;done",
						},
						Env:   map[string]string{"env": "env-value"},
						Image: "busybox",
						Metadata: map[string]string{
							cf.ProcessGUID: "thor",
							cf.LastUpdated: "4567.8",
							cf.VcapAppUris: "my.example.route",
							"routes":       "my.example.route",
						},
					},
				))
			})

			Context("when some have ready a ready revision", func() {
				BeforeEach(func() {
					l := createLRP("zeus", "irrelevant", "irrelevant")
					service := toService(l, namespace)
					service.Status = v1alpha1.ServiceStatus{LatestReadyRevisionName: "zeus-v1"}
					_, createErr := servingClient.Serving().Services(namespace).Create(service)
					Expect(createErr).ToNot(HaveOccurred())
					_, createErr = client.AppsV1beta2().ReplicaSets(namespace).Create(toReplicaSet(l, "zeus-v1", 2, 2, 1))
					Expect(createErr).ToNot(HaveOccurred())
				})

				It("sets the Target and Running fields", func() {
					actualLRPs, err := getter.List()
					Expect(err).ToNot(HaveOccurred())

					var lrp *opi.LRP
					for _, l := range actualLRPs {
						if l.Name == "zeus" {
							lrp = l
						}
					}
					Expect(lrp.TargetInstances).To(Equal(2))
					Expect(lrp.RunningInstances).To(Equal(1))
				})
			})
		})

		Context("When no replicaSet exists", func() {
			It("returns an empy list of LRPs", func() {
				actualLRPs, err := getter.List()
				Expect(err).ToNot(HaveOccurred())
				Expect(actualLRPs).To(BeEmpty())
			})
		})
	})
})

// FIXME - This is again reimplementing code
func toReplicaSet(lrp *opi.LRP, name string, t, c, r int32) *v1beta2.ReplicaSet {
	envs := MapToEnvVar(lrp.Env)
	envs = append(envs, v1.EnvVar{
		Name: "POD_NAME",
		ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{
				FieldPath: "metadata.name",
			},
		},
	})

	targetInstances := int32(lrp.TargetInstances)
	replicaSet := &v1beta2.ReplicaSet{
		Spec: v1beta2.ReplicaSetSpec{
			Replicas: &targetInstances,
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:    "opi",
							Image:   lrp.Image,
							Command: lrp.Command,
							Env:     envs,
							Ports: []v1.ContainerPort{
								{
									Name:          "expose",
									ContainerPort: 8080,
								},
							},
							LivenessProbe: &v1.Probe{},
						},
					},
				},
			},
		},
	}

	replicaSet.Name = name
	replicaSet.Namespace = namespace
	replicaSet.Spec.Selector = &meta.LabelSelector{
		MatchLabels: map[string]string{
			serving.RevisionLabelKey: lrp.Name,
		},
	}

	replicaSet.Labels = map[string]string{
		serving.RevisionLabelKey: name,
	}

	replicaSet.Status = v1beta2.ReplicaSetStatus{
		Replicas:          t,
		AvailableReplicas: c,
		ReadyReplicas:     r,
	}
	return replicaSet
}
