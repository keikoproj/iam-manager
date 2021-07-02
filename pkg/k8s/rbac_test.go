package k8s_test

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/keikoproj/iam-manager/pkg/k8s"
	mock_k8s "github.com/keikoproj/iam-manager/pkg/k8s/mocks"
	"gopkg.in/check.v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

type RBACSuite struct {
	t          *testing.T
	ctx        context.Context
	mockCtrl   *gomock.Controller
	mockClient *mock_k8s.MockClient
	mockRbac   *k8s.Client
}

func TestRBACTestSuite(t *testing.T) {
	check.Suite(&RBACSuite{t: t})
	check.TestingT(t)
}

func (s *RBACSuite) SetUpTest(c *check.C) {
	s.ctx = context.Background()
	s.mockCtrl = gomock.NewController(s.t)
	s.mockClient = mock_k8s.NewMockClient(s.mockCtrl)
	s.mockRbac = k8s.NewK8sManagerClient(s.mockClient)
}

func (s *RBACSuite) TearDownTest(c *check.C) {
	s.mockCtrl.Finish()
}

func (s *RBACSuite) TestEnsureServiceAccountAlreadyExistsAndIsCorrect(c *check.C) {
	req := k8s.ServiceAccountRequest{
		Namespace:          "test-ns",
		IamRoleARN:         "arn:aws:...",
		ServiceAccountName: "test-sa",
	}
	// Returning nil here triggers the behavior where a ServiceAccount object is indeed returned by the
	// GetServiceAccount function. However that object will be empty (it will have no annotations), so
	// we'll also be mocking out the
	s.mockClient.EXPECT().Get(s.ctx, client.ObjectKey{Namespace: "test-ns", Name: "test-sa"}, gomock.Any()).Times(1).Return(nil)
	err := s.mockRbac.EnsureServiceAccount(s.ctx, req)
	c.Assert(err, check.IsNil)
}

//############

// Failing:
//
// /private/var/folders/dm/b5by_qw91nd0ctdjbvggzgr40000gq/T/___RBACSuite_TestPatchServiceAccountAnnotation_in_github_com_keikoproj_iam_manager_pkg_k8s__3_.test -test.v -test.paniconexit0 -check.f ^\QTestPatchServiceAccountAnnotation\E$ -check.vv
// rbac.go:128: Unexpected call to *mock_k8s.MockClient.Patch([context.Background &ServiceAccount{ObjectMeta:{test-sa  test-ns    0 0001-01-01 00:00:00 +0000 UTC <nil> <nil> map[] map[] [] []  []},Secrets:[]ObjectReference{},ImagePullSecrets:[]LocalObjectReference{},AutomountServiceAccountToken:nil,} 0x1400012b5f0]) at /Users/diranged/go/src/github.com/keikoproj/iam-manager/pkg/k8s/rbac.go:128 because:
// expected call at /Users/diranged/go/src/github.com/keikoproj/iam-manager/pkg/k8s/rbac_test.go:49 doesn't match the argument at index 2.
// Got: &{application/strategic-merge-patch+json [123 34 109 101 116 97 100 97 116 97 34 58 123 34 97 110 110 111 116 97 116 105 111 110 115 34 58 123 34 102 97 107 101 45 97 110 110 111 116 97 116 105 111 110 34 58 32 34 118 97 108 117 101 34 125 125 125]} (*client.patch)
// Want: is equal to &{application/strategic-merge-patch+json [123 34 109 101 116 97 100 97 116 97 34 58 123 34 97 110 110 111 116 97 116 105 111 110 115 34 58 123 34 102 97 107 101 45 97 110 110 111 116 97 116 105 111 110 34 58 34 32 34 118 97 108 117 101 34 125 125 125]} (*client.patch)
// controller.go:269: missing call(s) to *mock_k8s.MockClient.Patch(is equal to context.Background (*context.emptyCtx), is anything, is equal to &{application/strategic-merge-patch+json [123 34 109 101 116 97 100 97 116 97 34 58 123 34 97 110 110 111 116 97 116 105 111 110 115 34 58 123 34 102 97 107 101 45 97 110 110 111 116 97 116 105 111 110 34 58 34 32 34 118 97 108 117 101 34 125 125 125]} (*client.patch)) /Users/diranged/go/src/github.com/keikoproj/iam-manager/pkg/k8s/rbac_test.go:49
// controller.go:269: aborting test due to missing call(s)
func (s *RBACSuite) TestPatchServiceAccountAnnotation(c *check.C) {
	patch := []byte(`{"metadata":{"annotations":{"fake-annotation":" "value"}}}`)
	s.mockClient.EXPECT().Patch(s.ctx, gomock.Any(), client.RawPatch(types.StrategicMergePatchType, patch)).Times(1).Return(nil)
	err := s.mockRbac.PatchServiceAccountAnnotation(s.ctx, "test-sa", "test-ns", "fake-annotation", "value")
	c.Assert(err, check.IsNil)

}
