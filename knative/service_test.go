package knative_test

import (
	"code.cloudfoundry.org/eirini"
	"code.cloudfoundry.org/eirini/opi"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "code.cloudfoundry.org/eirini/knative"
	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	servingclientset "github.com/knative/serving/pkg/client/clientset/versioned"
	"github.com/knative/serving/pkg/client/clientset/versioned/fake"
)

var _ = Describe("Service", func() {
	var (
		fakeClient     servingclientset.Interface
		serviceManager ServiceManager
		routesChan     chan []*eirini.Routes
	)

	const (
		namespace = "midgard"
	)

	BeforeEach(func() {
		routesChan = make(chan []*eirini.Routes, 1)
		fakeClient = fake.NewSimpleClientset()
		serviceManager = NewServiceManager(fakeClient, namespace, routesChan)
	})

	Context("When checking if a service exists", func() {
		var (
			exists bool
			name   string
			err    error
		)

		JustBeforeEach(func() {
			exists, err = serviceManager.Exists(name)
		})

		Context("when the service exists", func() {

			BeforeEach(func() {
				name = "baldur"
				lrp := createLRP(name, "9012.3", "my.example.route")
				service := toService(lrp, namespace)
				_, err := fakeClient.Serving().Services(namespace).Create(service)
				Expect(err).ToNot(HaveOccurred())
			})

			It("should not return an error", func() {
				Expect(err).ToNot(HaveOccurred())
			})

			It("shold return true", func() {
				Expect(exists).To(Equal(true))
			})
		})

		Context("when the service does not exist", func() {

			BeforeEach(func() {
				name = "non-existent"
			})

			It("should not return an error", func() {
				Expect(err).ToNot(HaveOccurred())
			})

			It("shold return true", func() {
				Expect(exists).To(Equal(false))
			})
		})
	})

	Context("When exposing an existing LRP", func() {

		var (
			lrp *opi.LRP
			err error
		)

		BeforeEach(func() {
			lrp = createLRP("baldur", "54321.0", `["my.example.route"]`)
		})

		Context("When creating a usual service", func() {

			JustBeforeEach(func() {
				err = serviceManager.Create(lrp)
			})

			It("should not fail", func() {
				Expect(err).ToNot(HaveOccurred())
			})

			It("should create a service", func() {
				serviceName := eirini.GetInternalServiceName("baldur")
				service, getErr := fakeClient.Serving().Services(namespace).Get(serviceName, meta.GetOptions{})
				Expect(getErr).ToNot(HaveOccurred())
				Expect(service).To(Equal(toService(lrp, namespace)))
			})

			It("should submit the routes to the route channel", func() {
				Eventually(routesChan).Should(Receive())
			})

			It("should emmit the correct routes", func() {
				routes := <-routesChan
				Expect(routes).To(HaveLen(1))

				route := routes[0]
				Expect(route.Routes).To(HaveLen(1))
				Expect(route.Routes).To(ContainElement("my.example.route"))
				Expect(route.UnregisteredRoutes).To(HaveLen(0))
			})

			Context("When recreating a existing service", func() {
				BeforeEach(func() {
					lrp = createLRP("baldur", "54321.0", `["my.example.route"]`)
				})

				JustBeforeEach(func() {
					err = serviceManager.Create(lrp)
				})

				It("should return an error", func() {
					Expect(err).To(HaveOccurred())
				})
			})
		})
	})

	Context("When deleting", func() {
		var service *v1alpha1.Service

		Context("a regular service", func() {

			var (
				err  error
				work []*eirini.Routes
			)

			BeforeEach(func() {
				lrp := createLRP("odin", "1234.5", `["my.example.route"]`)
				service = toService(lrp, namespace)
				_, err = fakeClient.Serving().Services(namespace).Create(service)
				Expect(err).ToNot(HaveOccurred())
			})

			JustBeforeEach(func() {
				err = serviceManager.Delete("odin")
			})

			It("should not return an error", func() {
				Expect(err).ToNot(HaveOccurred())
			})

			It("sends work to the route emitter", func() {
				Eventually(routesChan).Should(Receive())
			})

			It("should send the right routes to unregister", func() {
				work = <-routesChan
				Expect(work[0].UnregisteredRoutes).To(ContainElement("my.example.route"))
			})

			It("should delete the service", func() {
				_, err = fakeClient.Serving().Services(namespace).Get(service.Name, meta.GetOptions{})
				Expect(err).To(HaveOccurred())
			})

			Context("when the service does not exist", func() {

				JustBeforeEach(func() {
					err = serviceManager.Delete("tyr")
				})

				It("returns an error", func() {
					Expect(err).To(HaveOccurred())
				})
			})
		})
	})

	Context("When updating an service", func() {
		var (
			err            error
			lrp            *opi.LRP
			serviceName    string
			updatedService *v1alpha1.Service
		)

		BeforeEach(func() {
			lrp = createLRP("odin", "1234.5", `["my.example.route"]`)
			err = serviceManager.Create(lrp)
			r := <-routesChan
			Expect(r).To(HaveLen(1))
		})

		Context("when routes are updated", func() {

			JustBeforeEach(func() {
				err = serviceManager.Update(lrp)
				Expect(err).ToNot(HaveOccurred())

				serviceName = eirini.GetInternalServiceName("odin")

				updatedService, err = fakeClient.Serving().Services(namespace).Get(serviceName, meta.GetOptions{})
			})

			Context("When a route is replaced", func() {
				BeforeEach(func() {
					lrp = createLRP("odin", "1234.5", `["my-new.example.route"]`)
					lrp.Metadata["new-annotation"] = "update-me"
				})

				It("should not return an error", func() {
					Expect(err).ToNot(HaveOccurred())
				})

				It("should update the routes annotation", func() {
					Expect(updatedService.Annotations[eirini.RegisteredRoutes]).To(Equal(`["my-new.example.route"]`))
				})

				It("should update annotations", func() {
					Expect(updatedService.Annotations["new-annotation"]).To(Equal(`update-me`))
				})

				It("should remove the difference and emmit the routes", func() {
					routes := <-routesChan
					Expect(routes).To(HaveLen(1))

					route := routes[0]
					Expect(route.Routes).To(HaveLen(1))
					Expect(route.Routes).To(ContainElement("my-new.example.route"))
					Expect(route.UnregisteredRoutes).To(HaveLen(1))
					Expect(route.UnregisteredRoutes).To(ContainElement("my.example.route"))
				})
			})

			Context("When routes are added", func() {
				BeforeEach(func() {
					lrp = createLRP("odin", "1234.5", `["my.example.route","my-new.example.route"]`)
				})

				It("should contain the old route", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(updatedService.Annotations[eirini.RegisteredRoutes]).To(ContainSubstring(`"my.example.route"`))
				})

				It("should contain the new route", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(updatedService.Annotations[eirini.RegisteredRoutes]).To(ContainSubstring(`"my-new.example.route"`))
				})

				It("should remove the difference and emmit the routes", func() {
					routes := <-routesChan
					Expect(routes).To(HaveLen(1))

					route := routes[0]
					Expect(route.Routes).To(HaveLen(2))
					Expect(route.Routes).To(ContainElement("my-new.example.route"))
					Expect(route.Routes).To(ContainElement("my.example.route"))
					Expect(route.UnregisteredRoutes).To(HaveLen(0))
				})
			})

			Context("When routes are completly removed", func() {
				BeforeEach(func() {
					lrp = createLRP("odin", "1234.5", `[]`)
				})

				It("should empty the routes annotation", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(updatedService.Annotations[eirini.RegisteredRoutes]).To(Equal(`[]`))
				})

				It("should remove the difference and emmit the routes", func() {
					routes := <-routesChan
					Expect(routes).To(HaveLen(1))

					route := routes[0]
					Expect(route.UnregisteredRoutes).To(HaveLen(1))
					Expect(route.UnregisteredRoutes).To(ContainElement("my.example.route"))
					Expect(route.Routes).To(HaveLen(0))
				})
			})
		})
	})
})
