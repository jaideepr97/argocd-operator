package argocd

import (
	"context"
	"errors"
	"testing"

	"github.com/argoproj-labs/argocd-operator/api/v1alpha1"
	argoprojv1alpha1 "github.com/argoproj-labs/argocd-operator/api/v1alpha1"

<<<<<<< HEAD
=======
	oappsv1 "github.com/openshift/api/apps/v1"
	configv1 "github.com/openshift/api/config/v1"
>>>>>>> c8e4909 (fix: address CVE-2023-39325 (#1022))
	routev1 "github.com/openshift/api/route/v1"
	"github.com/stretchr/testify/assert"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

<<<<<<< HEAD
func TestReconcileArgoCD_reconcileStatusSSOConfig(t *testing.T) {
=======
func TestReconcileArgoCD_reconcileStatusKeycloak_K8s(t *testing.T) {
	logf.SetLogger(ZapLogger(true))

	a := makeTestArgoCDForKeycloak()

	resObjs := []client.Object{a}
	subresObjs := []client.Object{a}
	runtimeObjs := []runtime.Object{}
	sch := makeTestReconcilerScheme(argoproj.AddToScheme)
	cl := makeTestReconcilerClient(sch, resObjs, subresObjs, runtimeObjs)
	r := makeTestReconciler(cl, sch)

	assert.NoError(t, createNamespace(r, a.Namespace, ""))

	d := newKeycloakDeployment(a)

	// keycloak not installed
	_ = r.reconcileStatusKeycloak(a)
	assert.Equal(t, "Unknown", a.Status.SSO)

	// keycloak installation started
	r.Client.Create(context.TODO(), d)

	_ = r.reconcileStatusKeycloak(a)
	assert.Equal(t, "Pending", a.Status.SSO)

	// keycloak installation completed
	d.Status.ReadyReplicas = *d.Spec.Replicas
	r.Client.Status().Update(context.TODO(), d)

	_ = r.reconcileStatusKeycloak(a)
	assert.Equal(t, "Running", a.Status.SSO)
}

func TestReconcileArgoCD_reconcileStatusKeycloak_OpenShift(t *testing.T) {
	logf.SetLogger(ZapLogger(true))

	a := makeTestArgoCDForKeycloak()

	resObjs := []client.Object{a}
	subresObjs := []client.Object{a}
	runtimeObjs := []runtime.Object{}
	sch := makeTestReconcilerScheme(argoproj.AddToScheme)
	cl := makeTestReconcilerClient(sch, resObjs, subresObjs, runtimeObjs)
	r := makeTestReconciler(cl, sch)

	assert.NoError(t, createNamespace(r, a.Namespace, ""))

	assert.NoError(t, oappsv1.AddToScheme(r.Scheme))
	templateAPIFound = true
	defer removeTemplateAPI()

	dc := getKeycloakDeploymentConfigTemplate(a)
	dc.ObjectMeta.Name = defaultKeycloakIdentifier

	// keycloak not installed
	_ = r.reconcileStatusKeycloak(a)
	assert.Equal(t, "Unknown", a.Status.SSO)

	// create new client with dc object already present, but with 0 ready replicas to simulate
	// keycloak installation started
	resObjs = append(resObjs, dc)
	subresObjs = append(subresObjs, dc)
	r.Client = makeTestReconcilerClient(sch, resObjs, subresObjs, runtimeObjs)

	_ = r.reconcileStatusKeycloak(a)
	assert.Equal(t, "Pending", a.Status.SSO)

	// create new client with dc object already present, with 1 ready replica to simulate
	// keycloak installation completed
	dc.Status.ReadyReplicas = dc.Spec.Replicas
	r.Client = makeTestReconcilerClient(sch, resObjs, subresObjs, runtimeObjs)

	_ = r.reconcileStatusKeycloak(a)
	assert.Equal(t, "Running", a.Status.SSO)
}

func TestReconcileArgoCD_reconcileStatusSSO(t *testing.T) {
>>>>>>> c8e4909 (fix: address CVE-2023-39325 (#1022))
	logf.SetLogger(ZapLogger(true))

	tests := []struct {
		name             string
		argoCD           *argoprojv1alpha1.ArgoCD
		templateAPIfound bool
		wantSSOConfig    string
		wantErr          bool
		Err              error
	}{
		{
			name: "only dex configured",
			argoCD: makeTestArgoCD(func(ac *argoprojv1alpha1.ArgoCD) {
				ac.Spec.Dex = &argoprojv1alpha1.ArgoCDDexSpec{
					Resources:      makeTestDexResources(),
					OpenShiftOAuth: true,
				}
			}),
			templateAPIfound: false,
			wantSSOConfig:    "Success",
			wantErr:          false,
		},
		{
			name: "only keycloak configured",
			argoCD: makeTestArgoCD(func(cr *argoprojv1alpha1.ArgoCD) {
				cr.Spec.SSO = &v1alpha1.ArgoCDSSOSpec{
					Provider: argoprojv1alpha1.SSOProviderTypeKeycloak,
				}
				cr.Spec.Dex = &v1alpha1.ArgoCDDexSpec{
					OpenShiftOAuth: false,
				}
			}),
			templateAPIfound: true,
			wantSSOConfig:    "Success",
			wantErr:          false,
		},
		{
			name: "both dex and keycloak configured",
			argoCD: makeTestArgoCD(func(cr *argoprojv1alpha1.ArgoCD) {
				cr.Spec.SSO = &v1alpha1.ArgoCDSSOSpec{
					Provider: argoprojv1alpha1.SSOProviderTypeKeycloak,
				}
				cr.Spec.Dex = &v1alpha1.ArgoCDDexSpec{
					OpenShiftOAuth: true,
				}
			}),
			templateAPIfound: true,
			wantSSOConfig:    "Failed",
			wantErr:          true,
			Err:              errors.New("multiple SSO configuration"),
		},
		{
			name: "no sso configured",
			argoCD: makeTestArgoCD(func(cr *argoprojv1alpha1.ArgoCD) {
				cr.Spec.Dex = &v1alpha1.ArgoCDDexSpec{}
			}),
			templateAPIfound: false,
			wantSSOConfig:    "Unknown",
			wantErr:          false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			resObjs := []client.Object{test.argoCD}
			subresObjs := []client.Object{test.argoCD}
			runtimeObjs := []runtime.Object{}
			sch := makeTestReconcilerScheme(argoproj.AddToScheme)
			cl := makeTestReconcilerClient(sch, resObjs, subresObjs, runtimeObjs)
			r := makeTestReconciler(cl, sch)

			assert.NoError(t, createNamespace(r, test.argoCD.Namespace, ""))

			err := r.reconcileSSO(test.argoCD)

			err = r.reconcileStatusSSOConfig(test.argoCD)
			if err != nil {
				if !test.wantErr {
					t.Errorf("Got unexpected error")
				} else {
					assert.Equal(t, test.Err, err)
				}
			}

			assert.Equal(t, test.wantSSOConfig, test.argoCD.Status.SSOConfig)
		})
	}
}

func TestReconcileArgoCD_reconcileStatusHost(t *testing.T) {
	logf.SetLogger(ZapLogger(true))

	tests := []struct {
		name              string
		routeEnabled      bool
		testRouteAPIFound bool
		ingressEnabled    bool
		expectedNil       bool
		expectedHost      bool
		host              string
		phase             string
	}{
		{
			name:              "",
			routeEnabled:      true,
			testRouteAPIFound: true,
			ingressEnabled:    false,
			expectedNil:       false,
			host:              "argocd",
			phase:             "Available",
		},
		{
			name:              "",
			routeEnabled:      false,
			testRouteAPIFound: false,
			ingressEnabled:    true,
			expectedNil:       false,
			host:              "argocd, 12.0.0.5",
			phase:             "Available",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			routeAPIFound = test.testRouteAPIFound

			a := makeTestArgoCD(func(a *argoprojv1alpha1.ArgoCD) {
				a.Spec.Server.Route.Enabled = test.routeEnabled
				a.Spec.Server.Ingress.Enabled = test.ingressEnabled
			})

			route := &routev1.Route{
				ObjectMeta: metav1.ObjectMeta{
					Name:      testArgoCDName + "-server",
					Namespace: testNamespace,
					Labels: map[string]string{
						"app.kubernetes.io/name": testArgoCDName + "-server",
					},
				},
				Spec: routev1.RouteSpec{
					Host: "argocd",
				},
				Status: routev1.RouteStatus{
					Ingress: []routev1.RouteIngress{
						{
							Host: "argocd",
							Conditions: []routev1.RouteIngressCondition{
								{
									Type:   routev1.RouteAdmitted,
									Status: "True",
								},
							},
						},
					},
				},
			}

			ingress := &networkingv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:      testArgoCDName + "-server",
					Namespace: testNamespace,
				},
				Status: networkingv1.IngressStatus{
					LoadBalancer: networkingv1.IngressLoadBalancerStatus{
						Ingress: []networkingv1.IngressLoadBalancerIngress{
							{
								IP:       "12.0.0.1",
								Hostname: "argocd",
								Ports:    []networkingv1.IngressPortStatus{},
							},
							{
								IP:       "12.0.0.5",
								Hostname: "",
							},
						},
					},
				},
			}

			resObjs := []client.Object{a}
			subresObjs := []client.Object{a}
			runtimeObjs := []runtime.Object{}
			sch := makeTestReconcilerScheme(argoproj.AddToScheme, configv1.AddToScheme, routev1.AddToScheme)
			cl := makeTestReconcilerClient(sch, resObjs, subresObjs, runtimeObjs)
			r := makeTestReconciler(cl, sch)

			if test.routeEnabled {
				err := r.Client.Create(context.TODO(), route)
				assert.NoError(t, err)

			} else if test.ingressEnabled {
				err := r.Client.Create(context.TODO(), ingress)
				assert.NoError(t, err)
				assert.NotEqual(t, "Pending", a.Status.Phase)
			}

			err := r.reconcileStatusHost(a)
			assert.NoError(t, err)

			assert.Equal(t, test.host, a.Status.Host)
			assert.Equal(t, test.phase, a.Status.Phase)
		})
	}
}

func TestReconcileArgoCD_reconcileStatusNotificationsController(t *testing.T) {
	logf.SetLogger(ZapLogger(true))
	a := makeTestArgoCD()

	resObjs := []client.Object{a}
	subresObjs := []client.Object{a}
	runtimeObjs := []runtime.Object{}
	sch := makeTestReconcilerScheme(argoproj.AddToScheme)
	cl := makeTestReconcilerClient(sch, resObjs, subresObjs, runtimeObjs)
	r := makeTestReconciler(cl, sch)

	assert.NoError(t, r.reconcileStatusNotifications(a))
	assert.Equal(t, "", a.Status.NotificationsController)

	a.Spec.Notifications.Enabled = true
	assert.NoError(t, r.reconcileNotificationsController(a))
	assert.NoError(t, r.reconcileStatusNotifications(a))
	assert.Equal(t, "Pending", a.Status.NotificationsController)

	a.Spec.Notifications.Enabled = false
	assert.NoError(t, r.deleteNotificationsResources(a))
	assert.NoError(t, r.reconcileStatusNotifications(a))
	assert.Equal(t, "", a.Status.NotificationsController)
}

func TestReconcileArgoCD_reconcileStatusApplicationSetController(t *testing.T) {
	logf.SetLogger(ZapLogger(true))
	a := makeTestArgoCD()

	resObjs := []client.Object{a}
	subresObjs := []client.Object{a}
	runtimeObjs := []runtime.Object{}
	sch := makeTestReconcilerScheme(argoproj.AddToScheme)
	cl := makeTestReconcilerClient(sch, resObjs, subresObjs, runtimeObjs)
	r := makeTestReconciler(cl, sch)

	assert.NoError(t, r.reconcileStatusApplicationSetController(a))
	assert.Equal(t, "Unknown", a.Status.ApplicationSetController)

	a.Spec.ApplicationSet = &v1alpha1.ArgoCDApplicationSet{}
	assert.NoError(t, r.reconcileApplicationSetController(a))
	assert.NoError(t, r.reconcileStatusApplicationSetController(a))
	assert.Equal(t, "Pending", a.Status.ApplicationSetController)
}
