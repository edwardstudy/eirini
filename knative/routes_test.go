package knative_test

import (
	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	servingclientset "github.com/knative/serving/pkg/client/clientset/versioned"
	"github.com/knative/serving/pkg/client/clientset/versioned/fake"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/eirini"
	. "code.cloudfoundry.org/eirini/knative"
	"code.cloudfoundry.org/eirini/opi"
	"code.cloudfoundry.org/eirini/route"
)

const (
	namespace = "testing"
)

var _ = Describe("Routes", func() {
	Context("ListRoutes", func() {
		var (
			fakeClient  servingclientset.Interface
			routeLister route.Lister
			routes      []*eirini.Routes
			err         error
		)

		const kubeNamespace = "testing"

		BeforeEach(func() {
			fakeClient = fake.NewSimpleClientset()
			routeLister = NewServiceRouteLister(fakeClient, kubeNamespace)
		})

		JustBeforeEach(func() {
			routes, err = routeLister.ListRoutes()
		})

		Context("When there are existing services", func() {
			var lrp *opi.LRP

			BeforeEach(func() {
				lrp = createLRP("baldur", "54321.0", `["my.example.route"]`)
				_, err = fakeClient.Serving().Services(kubeNamespace).Create(toService(lrp, kubeNamespace))
				Expect(err).ToNot(HaveOccurred())
			})

			It("should not return an error", func() {
				Expect(err).ToNot(HaveOccurred())
			})

			It("should return the correct routes", func() {
				Expect(routes).To(HaveLen(1))
				route := routes[0]
				Expect(route.Routes).To(ContainElement("my.example.route"))
				Expect(route.Name).To(Equal(eirini.GetInternalServiceName("baldur")))
			})

			Context("When there are non cf services", func() {
				BeforeEach(func() {
					service := &v1alpha1.Service{}
					service.Name = "some-other-service"
					_, err = fakeClient.Serving().Services(namespace).Create(service)
					Expect(err).ToNot(HaveOccurred())
				})

				It("should not return an error", func() {
					Expect(err).ToNot(HaveOccurred())
				})

				It("should return only one Routes object", func() {
					Expect(routes).To(HaveLen(1))
				})
			})
		})
	})
})
