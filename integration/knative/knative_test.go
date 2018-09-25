package integration_test

// +build integration

import (
	"code.cloudfoundry.org/eirini"
	"code.cloudfoundry.org/eirini/knative"
	"code.cloudfoundry.org/eirini/models/cf"
	"code.cloudfoundry.org/eirini/opi"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Knative Desirer", func() {
	var (
		desirer    *knative.Desirer
		lrp        *opi.LRP
		err        error
		routesChan chan []*eirini.Routes
		appName    string
	)

	BeforeEach(func() {
		lrp = &opi.LRP{
			Name: "odin",
			Command: []string{
				"/bin/sh",
				"-c",
				"while true; do echo hello; sleep 10;done",
			},
			TargetInstances: 2,
			Image:           "busybox",
			Metadata: map[string]string{
				cf.ProcessGUID: "odin",
				cf.VcapAppUris: "[]",
			},
		}
		appName = eirini.GetInternalServiceName(lrp.Name)
		routesChan = make(chan []*eirini.Routes, 1)
		go func() {
			for {
				<-routesChan
			}
		}()
	})

	AfterEach(func() {
		if _, err := services().Get(appName, meta.GetOptions{}); err == nil {
			backgroundPropagation := meta.DeletePropagationBackground
			err = services().Delete(appName, &meta.DeleteOptions{PropagationPolicy: &backgroundPropagation})
			Expect(err).ToNot(HaveOccurred())
		}

		Eventually(listAllServices, timeout).Should(BeEmpty())

	})

	JustBeforeEach(func() {
		desirer = knative.NewDesirer(
			namespace,
			clientset,
			servingClient,
			routesChan,
		)
	})

	Context("When creating an LRP", func() {
		JustBeforeEach(func() {
			err = desirer.Desire(lrp)
		})

		It("should not fail", func() {
			Expect(err).ToNot(HaveOccurred())
		})

		It("should create a service object", func() {
			service, getErr := services().Get(appName, meta.GetOptions{})
			Expect(getErr).ToNot(HaveOccurred())

			Expect(service.Name).To(Equal(appName))
		})

		// TODO
		// It("should create associated revision", func() {
		// })

		// TODO
		// It("should create at least 1 pods", func() {
		// })

		// TODO
		// It("should create the associated route", func() {
		// })
	})

	Context("When stopping an LRP", func() {
		Context("when there is an existing LRP", func() {
			JustBeforeEach(func() {
				err = desirer.Desire(lrp)
				Expect(err).ToNot(HaveOccurred())
				Eventually(listAllServices, timeout).ShouldNot(BeEmpty())
				err = desirer.Stop("odin")
			})

			It("should not fail", func() {
				Expect(err).ToNot(HaveOccurred())
			})

			It("should delete the Service object", func() {
				Eventually(listAllServices, timeout).Should(BeEmpty())
			})

			// TODO
			// It("should delete the associated pods", func() {
			// })

			// TODO
			// It("should delete the associated route", func() {
			// })
		})

		Context("when the LRP does NOT exist", func() {
			JustBeforeEach(func() {
				err = desirer.Desire(lrp)
				Expect(err).ToNot(HaveOccurred())
				Eventually(listAllServices, timeout).ShouldNot(BeEmpty())
			})

			It("does something", func() {
				err = desirer.Stop("not-existing")
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Context("When getting an app", func() {
		JustBeforeEach(func() {
			err = desirer.Desire(lrp)
			Expect(err).ToNot(HaveOccurred())
		})

		It("correctly reports the running instances", func() {
			Eventually(func() int {
				l, e := desirer.Get("odin")
				Expect(e).ToNot(HaveOccurred())
				return l.RunningInstances
			}, timeout).Should(Equal(1))
		})

		Context("When one of the instances if failing", func() {
			BeforeEach(func() {
				lrp = &opi.LRP{
					Name: "odin",
					Command: []string{
						"/bin/sh",
						"-c",
						"exit;",
					},
					TargetInstances: 2,
					Image:           "busybox",
					Metadata: map[string]string{
						cf.VcapAppUris: "[]",
						cf.ProcessGUID: "odin",
					},
				}
			})

			It("correctly reports 1 target instances", func() {
				Eventually(func() int {
					lrp, err := desirer.Get("odin")
					Expect(err).NotTo(HaveOccurred())
					Expect(lrp.RunningInstances).To(BeZero())
					return lrp.TargetInstances
				}, timeout).Should(Equal(1))
			})
		})
	})
})

func int32ptr(i int) *int32 {
	i32 := int32(i)
	return &i32
}
