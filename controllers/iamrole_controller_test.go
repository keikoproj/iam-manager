package controllers_test

import (
	iammanagerv1alpha1 "github.com/keikoproj/iam-manager/api/v1alpha1"
	. "github.com/keikoproj/iam-manager/controllers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

var _ = Describe("IamroleController", func() {
	Describe("When checking a StatusUpdatePredicate", func() {
		instance := StatusUpdatePredicate{}

		Context("Where status update request made", func() {
			It("Should return false", func() {
				new := &iammanagerv1alpha1.Iamrole{
					Status: iammanagerv1alpha1.IamroleStatus{
						RoleName:   "role1",
						RetryCount: 2,
						State:      iammanagerv1alpha1.Error,
					},
				}

				old := &iammanagerv1alpha1.Iamrole{
					Status: iammanagerv1alpha1.IamroleStatus{
						RoleName:   "role1",
						RetryCount: 1,
						State:      iammanagerv1alpha1.Error,
					},
				}
				failEvt1 := event.UpdateEvent{MetaOld: old.GetObjectMeta(), ObjectOld: old, MetaNew: new.GetObjectMeta(), ObjectNew: new}
				failEvt2 := event.UpdateEvent{MetaOld: nil, ObjectOld: old, MetaNew: new.GetObjectMeta(), ObjectNew: new}
				failEvt3 := event.UpdateEvent{MetaOld: old.GetObjectMeta(), ObjectOld: nil, MetaNew: new.GetObjectMeta(), ObjectNew: new}
				failEvt4 := event.UpdateEvent{MetaOld: old.GetObjectMeta(), ObjectOld: old, MetaNew: nil, ObjectNew: new}
				failEvt5 := event.UpdateEvent{MetaOld: old.GetObjectMeta(), ObjectOld: old, MetaNew: new.GetObjectMeta(), ObjectNew: nil}

				Expect(instance.Update(failEvt1)).To(BeFalse())
				Expect(instance.Update(failEvt2)).To(BeFalse())
				Expect(instance.Update(failEvt3)).To(BeFalse())
				Expect(instance.Update(failEvt4)).To(BeFalse())
				Expect(instance.Update(failEvt5)).To(BeFalse())

			})
		})

		Context("Where status create request made", func() {
			It("Should return true", func() {

				Expect(instance.Create(event.CreateEvent{})).To(BeTrue())
			})
		})

		Context("Where status delete request made", func() {
			It("Should return true", func() {

				Expect(instance.Delete(event.DeleteEvent{})).To(BeTrue())
			})
		})

	})
})
