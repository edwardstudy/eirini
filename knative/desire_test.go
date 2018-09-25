package knative_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "code.cloudfoundry.org/eirini/knative"
	"code.cloudfoundry.org/eirini/knative/knativefakes"
	"code.cloudfoundry.org/eirini/opi"
)

var _ = Describe("Desire", func() {
	var getter *knativefakes.FakeServiceGetter
	var manager *knativefakes.FakeServiceManager
	var desirer *Desirer

	BeforeEach(func() {
		getter = new(knativefakes.FakeServiceGetter)
		manager = new(knativefakes.FakeServiceManager)
		desirer = &Desirer{ServiceGetter: getter, ServiceManager: manager}
	})

	Context("Desire", func() {
		var err error
		var lrp *opi.LRP

		JustBeforeEach(func() {
			lrp = &opi.LRP{Name: "odin"}
			err = desirer.Desire(lrp)
		})

		Context("when service does exist", func() {
			BeforeEach(func() {
				manager.ExistsReturns(true, nil)
			})

			It("does NOT call the manager", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(manager.CreateCallCount()).To(BeZero())
			})
		})

		Context("when checking if service exists fails", func() {
			BeforeEach(func() {
				manager.ExistsReturns(false, errors.New("didnt work"))
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when the service does not exist", func() {
			BeforeEach(func() {
				manager.ExistsReturns(false, nil)
			})

			It("calls the ServiceManager", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(manager.CreateCallCount()).To(Equal(1))
				Expect(manager.CreateArgsForCall(0)).To(Equal(lrp))
			})

			Context("when creating service fails", func() {
				BeforeEach(func() {
					manager.CreateReturns(errors.New("didnt work"))
				})

				It("returns an error", func() {
					Expect(err).To(HaveOccurred())
				})
			})
		})
	})

	Context("List", func() {
		var err error
		var got []*opi.LRP
		var want []*opi.LRP

		BeforeEach(func() {
			want = []*opi.LRP{&opi.LRP{Name: "odin"}}
			getter.ListReturns(want, nil)
		})

		JustBeforeEach(func() {
			got, err = desirer.List()
		})

		It("calls the ServiceGetter", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(getter.ListCallCount()).To(Equal(1))
			Expect(got).To(Equal(want))
		})

		Context("when getting the list of services fails", func() {
			BeforeEach(func() {
				getter.ListReturns(nil, errors.New("didnt work"))
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Context("Get", func() {
		var err error
		var got *opi.LRP
		var want *opi.LRP

		BeforeEach(func() {
			want = &opi.LRP{Name: "odin"}
			getter.GetReturns(want, nil)
		})

		JustBeforeEach(func() {
			got, err = desirer.Get("odin")
		})

		It("calls the ServiceGetter", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(getter.GetCallCount()).To(Equal(1))
			Expect(got).To(Equal(want))
		})

		Context("when getting the service fails", func() {
			BeforeEach(func() {
				getter.GetReturns(nil, errors.New("didnt work"))
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Context("Update", func() {
		var err error
		var want *opi.LRP

		BeforeEach(func() {
			want = &opi.LRP{Name: "odin"}
			getter.GetReturns(want, nil)
		})

		JustBeforeEach(func() {
			err = desirer.Update(want)
		})

		It("calls the ServiceManager", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(manager.UpdateCallCount()).To(Equal(1))
			Expect(manager.UpdateArgsForCall(0)).To(Equal(want))
		})

		Context("when getting the service fails", func() {
			BeforeEach(func() {
				manager.UpdateReturns(errors.New("didnt work"))
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Context("Stop", func() {
		var err error

		JustBeforeEach(func() {
			err = desirer.Stop("odin")
		})

		It("calls the ServiceManager", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(manager.DeleteCallCount()).To(Equal(1))
			Expect(manager.DeleteArgsForCall(0)).To(Equal("odin"))
		})

		Context("when stopping the service fails", func() {
			BeforeEach(func() {
				manager.DeleteReturns(errors.New("didnt work"))
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
