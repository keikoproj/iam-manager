package k8s_test

import (
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/keikoproj/iam-manager/constants"
	"github.com/keikoproj/iam-manager/pkg/k8s"
	"gopkg.in/check.v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	fakeclient "k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
	"testing"
)

type RBACSuite struct {
	t        *testing.T
	ctx      context.Context
	mockCtrl *gomock.Controller
	mockRbac *k8s.Client
}

func TestRBACTestSuite(t *testing.T) {
	check.Suite(&RBACSuite{t: t})
	check.TestingT(t)
}

func (s *RBACSuite) SetUpTest(c *check.C) {
	s.ctx = context.Background()
	s.mockCtrl = gomock.NewController(s.t)
}

func (s *RBACSuite) TearDownTest(c *check.C) {
	s.mockCtrl.Finish()
}

// ########## EnsureServiceAccount() Tests ##########

func (s *RBACSuite) TestEnsureServiceAccount(c *check.C) {
	mockClient := fakeclient.NewSimpleClientset()
	mockRbac := &k8s.Client{Cl: mockClient}
	req := k8s.ServiceAccountRequest{
		IamRoleARN:         "arn:...",
		ServiceAccountName: "test-sa",
		Namespace:          "default",
	}
	sa, err := mockRbac.EnsureServiceAccount(s.ctx, req)
	c.Assert(err, check.IsNil)
	c.Assert(sa.Name, check.Equals, "test-sa")
	c.Assert(sa.Namespace, check.Equals, "default")
	c.Assert(sa.Annotations[constants.ServiceAccountRoleAnnotation], check.Equals, "arn:...")
}

func (s *RBACSuite) TestEnsureServiceAccountWithUnexpectedCreateError(c *check.C) {
	mockClient := fakeclient.NewSimpleClientset()
	mockClient.PrependReactor(
		"create",
		"serviceaccounts",
		func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
			return true, &v1.ServiceAccount{}, errors.New("Error creating ServiceAccount")
		})
	mockRbac := &k8s.Client{Cl: mockClient}
	req := k8s.ServiceAccountRequest{
		IamRoleARN:         "arn:...",
		ServiceAccountName: "test-sa",
		Namespace:          "default",
	}
	sa, err := mockRbac.EnsureServiceAccount(s.ctx, req)
	c.Assert(err.Error(), check.Matches, "Error creating ServiceAccount")
	c.Assert(sa, check.IsNil)
}

func (s *RBACSuite) TestEnsureServiceAccountAlreadyExists(c *check.C) {
	mockClient := fakeclient.NewSimpleClientset(&v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test-sa",
			Namespace:   "default",
			Annotations: map[string]string{},
		},
	})
	mockRbac := &k8s.Client{Cl: mockClient}
	req := k8s.ServiceAccountRequest{
		IamRoleARN:         "arn:...",
		ServiceAccountName: "test-sa",
		Namespace:          "default",
	}
	sa, err := mockRbac.EnsureServiceAccount(s.ctx, req)
	c.Assert(err, check.IsNil)
	c.Assert(sa.Name, check.Equals, "test-sa")
	c.Assert(sa.Namespace, check.Equals, "default")
	c.Assert(sa.Annotations[constants.ServiceAccountRoleAnnotation], check.Equals, "arn:...")
}

func (s *RBACSuite) TestEnsureServiceAccountAlreadyExistsWithMatchingAnnotationButInvalidValue(c *check.C) {
	mockClient := fakeclient.NewSimpleClientset(&v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-sa",
			Namespace: "default",
			Annotations: map[string]string{
				constants.ServiceAccountRoleAnnotation: "arn:invaliddata",
			},
		},
	})
	mockRbac := &k8s.Client{Cl: mockClient}
	req := k8s.ServiceAccountRequest{
		IamRoleARN:         "arn:...",
		ServiceAccountName: "test-sa",
		Namespace:          "default",
	}
	sa, err := mockRbac.EnsureServiceAccount(s.ctx, req)
	c.Assert(err, check.IsNil)
	c.Assert(sa.Name, check.Equals, "test-sa")
	c.Assert(sa.Namespace, check.Equals, "default")
	c.Assert(sa.Annotations[constants.ServiceAccountRoleAnnotation], check.Equals, "arn:...")
}

func (s *RBACSuite) TestEnsureServiceAccountAlreadyExistsWithMatchingAnnotation(c *check.C) {
	mockClient := fakeclient.NewSimpleClientset(&v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-sa",
			Namespace: "default",
			Annotations: map[string]string{
				constants.ServiceAccountRoleAnnotation: "arn:...",
			},
		},
	})
	mockRbac := &k8s.Client{Cl: mockClient}
	req := k8s.ServiceAccountRequest{
		IamRoleARN:         "arn:...",
		ServiceAccountName: "test-sa",
		Namespace:          "default",
	}
	sa, err := mockRbac.EnsureServiceAccount(s.ctx, req)
	c.Assert(err, check.IsNil)
	c.Assert(sa.Name, check.Equals, "test-sa")
	c.Assert(sa.Namespace, check.Equals, "default")
	c.Assert(sa.Annotations[constants.ServiceAccountRoleAnnotation], check.Equals, "arn:...")
}

// ########## GetOrCreateServiceAccount() Tests ##########

func (s *RBACSuite) TestGetOrCreateServiceAccount(c *check.C) {
	mockClient := fakeclient.NewSimpleClientset()
	mockRbac := &k8s.Client{Cl: mockClient}
	sa, err := mockRbac.GetOrCreateServiceAccount(s.ctx, "test-sa", "default")
	c.Assert(err, check.IsNil)
	c.Assert(sa.Name, check.Equals, "test-sa")
}

func (s *RBACSuite) TestGetOrCreateServiceAccountAlreadyExists(c *check.C) {
	mockClient := fakeclient.NewSimpleClientset(&v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test-sa",
			Namespace:   "default",
			Annotations: map[string]string{},
		},
	})
	mockRbac := &k8s.Client{Cl: mockClient}
	sa, err := mockRbac.GetOrCreateServiceAccount(s.ctx, "test-sa", "default")
	c.Assert(err, check.IsNil)
	c.Assert(sa.Name, check.Equals, "test-sa")
}

// ########## GetServiceAccount() Tests #########

func (s *RBACSuite) TestGetServiceAccount(c *check.C) {
	mockClient := fakeclient.NewSimpleClientset(&v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test-sa",
			Namespace:   "default",
			Annotations: map[string]string{},
		},
	})
	mockRbac := &k8s.Client{Cl: mockClient}
	sa, err := mockRbac.GetServiceAccount(s.ctx, "test-sa", "default")
	c.Assert(err, check.IsNil)
	c.Assert(sa.Name, check.Equals, "test-sa")
}

func (s *RBACSuite) TestGetServiceAccountMissing(c *check.C) {
	mockClient := fakeclient.NewSimpleClientset()
	mockRbac := &k8s.Client{Cl: mockClient}
	sa, err := mockRbac.GetServiceAccount(s.ctx, "test-sa", "default")
	c.Assert(err, check.NotNil)
	c.Assert(sa, check.IsNil)
}

// ########## CreateServiceAccount() Tests ##########

func (s *RBACSuite) TestCreateServiceAccount(c *check.C) {
	mockClient := fakeclient.NewSimpleClientset()
	mockRbac := &k8s.Client{Cl: mockClient}
	sa, err := mockRbac.CreateServiceAccount(s.ctx, "test-sa", "default")
	c.Assert(err, check.IsNil)
	c.Assert(sa.Name, check.Equals, "test-sa")
}

func (s *RBACSuite) TestCreateServiceAccountAlreadyExists(c *check.C) {
	mockClient := fakeclient.NewSimpleClientset(&v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test-sa",
			Namespace:   "default",
			Annotations: map[string]string{},
		},
	})
	mockRbac := &k8s.Client{Cl: mockClient}
	sa, err := mockRbac.CreateServiceAccount(s.ctx, "test-sa", "default")
	c.Assert(err.Error(), check.Matches, "serviceaccounts \"test-sa\" already exists")
	c.Assert(sa, check.IsNil)
}

// ########## PatchServiceAccountAnnotation() Tests ##########

func (s *RBACSuite) TestPatchServiceAccountAnnotation(c *check.C) {
	mockClient := fakeclient.NewSimpleClientset(&v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-sa",
			Namespace: "default",
			Annotations: map[string]string{
				"some-other": "annotation",
			},
		},
	})
	mockRbac := &k8s.Client{Cl: mockClient}
	sa, err := mockRbac.PatchServiceAccountAnnotation(s.ctx, "test-sa", "default", "fake-annotation", "value")
	c.Assert(err, check.IsNil)
	c.Assert(sa.Annotations["fake-annotation"], check.Equals, "value")
	c.Assert(sa.Annotations["some-other"], check.Equals, "annotation")
}
