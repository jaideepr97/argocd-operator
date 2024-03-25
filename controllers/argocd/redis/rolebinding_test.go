package redis

import (
	"testing"

	argoproj "github.com/argoproj-labs/argocd-operator/api/v1beta1"
	"github.com/argoproj-labs/argocd-operator/controllers/argocd/argocdcommon"
	"github.com/argoproj-labs/argocd-operator/pkg/permissions"
	"github.com/argoproj-labs/argocd-operator/tests/test"
	"github.com/stretchr/testify/assert"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestReconcileRoleBinding(t *testing.T) {
	tests := []struct {
		name                string
		reconciler          *RedisReconciler
		expectedError       bool
		expectedRoleBinding *rbacv1.RoleBinding
	}{
		{
			name: "RoleBinding does not exist",
			reconciler: makeTestRedisReconciler(
				test.MakeTestArgoCD(nil),
			),
			expectedError:       false,
			expectedRoleBinding: getDesiredRoleBinding(),
		},
		{
			name: "RoleBinding does not exist, HA role",
			reconciler: makeTestRedisReconciler(
				test.MakeTestArgoCD(nil,
					func(ac *argoproj.ArgoCD) {
						ac.Spec.HA.Enabled = true
					},
				),
			),
			expectedError: false,
			expectedRoleBinding: test.MakeTestRoleBinding(getDesiredRoleBinding(),
				func(rb *rbacv1.RoleBinding) {
					rb.RoleRef = rbacv1.RoleRef{
						Kind:     "Role",
						Name:     "test-argocd-redis-ha",
						APIGroup: "rbac.authorization.k8s.io",
					}
				},
			),
		},
		{
			name: "RoleBinding drift",
			reconciler: makeTestRedisReconciler(
				test.MakeTestArgoCD(nil),
				test.MakeTestRoleBinding(getDesiredRoleBinding(),
					func(rb *rbacv1.RoleBinding) {
						rb.Name = "test-argocd-redis"
						// Modify some fields to simulate drift
						rb.Subjects = []rbacv1.Subject{
							{
								Kind:      "User",
								Name:      "test-user",
								Namespace: "test-namespace",
							},
						}
					},
				),
			),
			expectedError:       false,
			expectedRoleBinding: getDesiredRoleBinding(),
		},
		{
			name: "RoleBinding RoleRef drift",
			reconciler: makeTestRedisReconciler(
				test.MakeTestArgoCD(nil),
				test.MakeTestRoleBinding(getDesiredRoleBinding(),
					func(rb *rbacv1.RoleBinding) {
						rb.Name = "test-argocd-redis"
						// Modify RoleRef to simulate drift
						rb.RoleRef.Name = "test-different-role"
					},
				),
			),
			expectedError:       true,
			expectedRoleBinding: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.reconciler.varSetter()

			err := tt.reconciler.reconcileRoleBinding()
			assert.NoError(t, err)

			existing, err := permissions.GetRoleBinding("test-argocd-redis", test.TestNamespace, tt.reconciler.Client)

			if tt.expectedError {
				assert.Error(t, err, "Expected an error but got none.")
			} else {
				assert.NoError(t, err, "Expected no error but got one.")
			}

			if tt.expectedRoleBinding != nil {
				match := true

				// Check for partial match on relevant fields
				ftc := []argocdcommon.FieldToCompare{
					{
						Existing: existing.Labels,
						Desired:  tt.expectedRoleBinding.Labels,
					},
					{
						Existing: existing.Subjects,
						Desired:  tt.expectedRoleBinding.Subjects,
					},
					{
						Existing: existing.RoleRef,
						Desired:  tt.expectedRoleBinding.RoleRef,
					},
				}
				argocdcommon.PartialMatch(ftc, &match)
				assert.True(t, match)
			}
		})
	}
}

func TestDeleteRoleBinding(t *testing.T) {
	tests := []struct {
		name             string
		reconciler       *RedisReconciler
		roleBindingExist bool
		expectedError    bool
	}{
		{
			name: "RoleBinding exists",
			reconciler: makeTestRedisReconciler(
				test.MakeTestArgoCD(nil),
				test.MakeTestRoleBinding(nil),
			),
			roleBindingExist: true,
			expectedError:    false,
		},
		{
			name: "RoleBinding does not exist",
			reconciler: makeTestRedisReconciler(
				test.MakeTestArgoCD(nil),
			),
			roleBindingExist: false,
			expectedError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := tt.reconciler.deleteRoleBinding(test.TestName, test.TestNamespace)

			if tt.roleBindingExist {
				_, err := permissions.GetRoleBinding(test.TestName, test.TestNamespace, tt.reconciler.Client)
				assert.True(t, apierrors.IsNotFound(err))
			}

			if tt.expectedError {
				assert.Error(t, err, "Expected an error but got none.")
			} else {
				assert.NoError(t, err, "Expected no error but got one.")
			}
		})
	}
}

func getDesiredRoleBinding() *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-argocd-redis",
			Namespace: test.TestNamespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       "test-argocd-redis",
				"app.kubernetes.io/part-of":    "argocd",
				"app.kubernetes.io/instance":   "test-argocd",
				"app.kubernetes.io/managed-by": "argocd-operator",
				"app.kubernetes.io/component":  "redis",
			},
			Annotations: map[string]string{
				"argocds.argoproj.io/name":      "test-argocd",
				"argocds.argoproj.io/namespace": "test-ns",
			},
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "test-argocd-redis",
				Namespace: "test-ns",
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "Role",
			Name:     "test-argocd-redis",
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
}