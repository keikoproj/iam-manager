package controllers

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"

	iammanagerv1alpha1 "github.com/keikoproj/iam-manager/api/v1alpha1"
	"github.com/keikoproj/iam-manager/internal/config"
)

// roleWith returns an Iamrole with the given generation and annotation values.
// Empty annotation strings produce an absent annotation, which matches what
// the controller reads (Go's zero value on a missing map key is also "").
func roleWith(generation int64, irsa, tags string, status iammanagerv1alpha1.IamroleStatus) *iammanagerv1alpha1.Iamrole {
	annotations := map[string]string{}
	if irsa != "" {
		annotations[config.IRSAAnnotation] = irsa
	}
	if tags != "" {
		annotations[config.IamManagerTagsAnnotation] = tags
	}
	return &iammanagerv1alpha1.Iamrole{
		ObjectMeta: metav1.ObjectMeta{
			Generation:  generation,
			Annotations: annotations,
		},
		Status: status,
	}
}

func readyStatus(generation int64, irsa, tags string) iammanagerv1alpha1.IamroleStatus {
	return iammanagerv1alpha1.IamroleStatus{
		State:                  iammanagerv1alpha1.Ready,
		ObservedGeneration:     generation,
		ObservedIRSAAnnotation: irsa,
		ObservedTagsAnnotation: tags,
	}
}

// TestShouldShortCircuitReady exercises the truth table for Change 1. The
// helper is pure (no I/O, no globals beyond the explicit flag), so we can
// test every input combination directly without a fake client or envtest.
func TestShouldShortCircuitReady(t *testing.T) {
	tests := []struct {
		name    string
		role    *iammanagerv1alpha1.Iamrole
		enabled bool
		want    bool
	}{
		{
			name:    "flag off, inputs match — never short-circuit",
			role:    roleWith(3, "sa-foo", "k=v", readyStatus(3, "sa-foo", "k=v")),
			enabled: false,
			want:    false,
		},
		{
			name:    "flag on, all inputs match — short-circuit",
			role:    roleWith(3, "sa-foo", "k=v", readyStatus(3, "sa-foo", "k=v")),
			enabled: true,
			want:    true,
		},
		{
			name:    "generation bumped by spec write — must reconcile",
			role:    roleWith(4, "sa-foo", "k=v", readyStatus(3, "sa-foo", "k=v")),
			enabled: true,
			want:    false,
		},
		{
			name:    "IRSA annotation changed — must reconcile",
			role:    roleWith(3, "sa-bar", "k=v", readyStatus(3, "sa-foo", "k=v")),
			enabled: true,
			want:    false,
		},
		{
			name:    "IRSA annotation added — must reconcile",
			role:    roleWith(3, "sa-foo", "k=v", readyStatus(3, "", "k=v")),
			enabled: true,
			want:    false,
		},
		{
			name:    "IRSA annotation removed — must reconcile",
			role:    roleWith(3, "", "k=v", readyStatus(3, "sa-foo", "k=v")),
			enabled: true,
			want:    false,
		},
		{
			name:    "Tags annotation changed — must reconcile",
			role:    roleWith(3, "sa-foo", "k=v2", readyStatus(3, "sa-foo", "k=v")),
			enabled: true,
			want:    false,
		},
		{
			name:    "Fresh CR with no observed fields (generation=1, observed=0) — must reconcile",
			role:    roleWith(1, "", "", iammanagerv1alpha1.IamroleStatus{State: iammanagerv1alpha1.Ready}),
			enabled: true,
			want:    false,
		},
		{
			name:    "No IRSA / tags annotations and observed empty — short-circuit",
			role:    roleWith(3, "", "", readyStatus(3, "", "")),
			enabled: true,
			want:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := shouldShortCircuitReady(tc.role, tc.enabled)
			if got != tc.want {
				t.Fatalf("shouldShortCircuitReady() = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestStatusUpdatePredicate_Update_ResourceVersionDrop covers Change 2: when
// the predicate flag is enabled, Update events with an unchanged
// ResourceVersion are dropped before they enter the workqueue. When the flag
// is off the existing behavior is preserved.
func TestStatusUpdatePredicate_Update_ResourceVersionDrop(t *testing.T) {
	// Each subtest sets Props for the duration of the test. We restore the
	// previous value via t.Cleanup so other tests in the package aren't
	// affected by leakage.
	tests := []struct {
		name        string
		flagEnabled bool
		oldRV       string
		newRV       string
		// statusEqual controls whether old/new have identical Status. When
		// they do, the existing predicate returns true; when they differ it
		// returns false. We use this to make sure the new RV-equal branch
		// fires *before* the status-equality check.
		statusEqual bool
		want        bool
	}{
		{
			name:        "flag off, RV equal, status equal — existing path returns true",
			flagEnabled: false,
			oldRV:       "100",
			newRV:       "100",
			statusEqual: true,
			want:        true,
		},
		{
			name:        "flag off, RV equal, status differs — existing path returns false",
			flagEnabled: false,
			oldRV:       "100",
			newRV:       "100",
			statusEqual: false,
			want:        false,
		},
		{
			name:        "flag on, RV equal — drop regardless of status",
			flagEnabled: true,
			oldRV:       "100",
			newRV:       "100",
			statusEqual: true,
			want:        false,
		},
		{
			name:        "flag on, RV equal, status differs — still drop (early return)",
			flagEnabled: true,
			oldRV:       "100",
			newRV:       "100",
			statusEqual: false,
			want:        false,
		},
		{
			name:        "flag on, RV differs, status equal — fall through to existing rule",
			flagEnabled: true,
			oldRV:       "100",
			newRV:       "101",
			statusEqual: true,
			want:        true,
		},
		{
			name:        "flag on, RV differs, status differs — existing rule returns false",
			flagEnabled: true,
			oldRV:       "100",
			newRV:       "101",
			statusEqual: false,
			want:        false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			withResyncPredicateFlag(t, tc.flagEnabled)

			oldStatus := iammanagerv1alpha1.IamroleStatus{State: iammanagerv1alpha1.Ready, RoleName: "role1"}
			newStatus := oldStatus
			if !tc.statusEqual {
				newStatus.RetryCount = oldStatus.RetryCount + 1
			}

			oldObj := &iammanagerv1alpha1.Iamrole{
				ObjectMeta: metav1.ObjectMeta{ResourceVersion: tc.oldRV},
				Status:     oldStatus,
			}
			newObj := &iammanagerv1alpha1.Iamrole{
				ObjectMeta: metav1.ObjectMeta{ResourceVersion: tc.newRV},
				Status:     newStatus,
			}

			got := StatusUpdatePredicate{}.Update(event.UpdateEvent{ObjectOld: oldObj, ObjectNew: newObj})
			if got != tc.want {
				t.Fatalf("StatusUpdatePredicate.Update() = %v, want %v", got, tc.want)
			}
		})
	}
}

// withResyncPredicateFlag swaps config.Props for one whose
// IsResyncPredicateEnabled() returns `enabled`. We have to mutate the global
// because the predicate reads it directly; the t.Cleanup hook restores the
// previous value so tests can run in any order. The configmap below is the
// minimum that satisfies LoadProperties without it reaching out to AWS STS
// or EKS (both gated on empty aws.accountId / IRSA enabled).
func withResyncPredicateFlag(t *testing.T, enabled bool) {
	t.Helper()
	prev := config.Props
	t.Cleanup(func() { config.Props = prev })

	cm := &v1.ConfigMap{
		Data: map[string]string{
			"aws.accountId":                       "123456789012",
			"controller.resync.predicate.enabled": boolStr(enabled),
		},
	}
	if err := config.LoadProperties("", cm); err != nil {
		t.Fatalf("LoadProperties failed: %v", err)
	}
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
