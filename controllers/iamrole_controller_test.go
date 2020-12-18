package controllers_test

import (
	"context"
	"github.com/golang/mock/gomock"
	iammanagerv1alpha1 "github.com/keikoproj/iam-manager/api/v1alpha1"
	. "github.com/keikoproj/iam-manager/controllers"
	. "github.com/keikoproj/iam-manager/internal/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/check.v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"testing"
)

type ControllerSuite struct {
	t        *testing.T
	ctx      context.Context
	mockCtrl *gomock.Controller
}

func TestControllerSuite(t *testing.T) {
	check.Suite(&ControllerSuite{t: t})
	check.TestingT(t)
}

func (s *ControllerSuite) SetUpTest(c *check.C) {
	s.ctx = context.Background()
	s.mockCtrl = gomock.NewController(s.t)
}

func (s *ControllerSuite) TearDownTest(c *check.C) {
	s.mockCtrl.Finish()
}

func (s *ControllerSuite) TestGenerateNameFunction(c *check.C) {
	cm := &v1.ConfigMap{
		Data: map[string]string{
			"iam.role.derive.from.namespace": "false",
			"iam.role.prefix":                "pfx",
			"iam.role.separator":             "+",
		},
	}
	Props = nil
	err := LoadProperties("", cm)
	c.Assert(err, check.Equals, nil)

	resource := &iammanagerv1alpha1.Iamrole{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "test-ns",
		},
	}
	c.Assert(GenerateRoleName(*resource, *Props), check.Equals, "pfx+foo")
}

func (s *ControllerSuite) TestGenerateNameFunctionWithDeriveFromNamespaceEnabled(c *check.C) {
	cm := &v1.ConfigMap{
		Data: map[string]string{
			"iam.role.derive.from.namespace": "true",
			"iam.role.prefix":                "pfx",
			"iam.role.separator":             "+",
		},
	}
	Props = nil
	err := LoadProperties("", cm)
	c.Assert(err, check.Equals, nil)

	resource := &iammanagerv1alpha1.Iamrole{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "test-ns",
		},
	}
	c.Assert(GenerateRoleName(*resource, *Props), check.Equals, "pfx+test-ns")
}

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
