/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package profile

import (
	"context"
	"fmt"
	"github.com/kubeflow/kubeflow/v2/internal/controller/reconciler/profile/metrics"
	"strings"

	xpevent "github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/kubeflow/kubeflow/v2/apis/core/v1"
	"github.com/kubeflow/kubeflow/v2/internal/controller/reconciler/profile/watch"
	"github.com/kubeflow/kubeflow/v2/internal/logging"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	ProfileQuota  = "kf-resource-quota"
	DefaultEditor = "default-editor"
	DefaultViewer = "default-viewer"
)

var (
	labelOwner = fmt.Sprintf("%s/owner", v1.Group)
	labelRole  = fmt.Sprintf("%s/role", v1.Group)
)

func Setup(mgr ctrl.Manager, o controller.Options) error {

	name := "kubeflow/" + strings.ToLower(v1.ProfileKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1.Profile{}).
		Owns(&corev1.Namespace{}).
		Owns(&rbacv1.RoleBinding{}).
		Owns(&corev1.ResourceQuota{}).
		Owns(&corev1.ServiceAccount{}).
		Watches(
			&source.Kind{Type: &v1.ProfileConfig{}},
			&watch.EnqueueRequestForProfiles{Reader: mgr.GetClient()},
		).
		WithOptions(o).
		Complete(NewReconciler(mgr,
			WithLogger(logging.NewLogrLogger(mgr.GetLogger().WithValues("controller", name))),
			WithMetricsRecorder(metrics.NewPrometheusRecorder()),
		))

}

type ReconcilerOption func(r *Reconciler)

func WithLogger(log logging.Logger) ReconcilerOption {
	return func(r *Reconciler) {
		r.logger = log
	}
}

func WithMetricsRecorder(m metrics.Recorder) ReconcilerOption {
	return func(r *Reconciler) {
		r.metrics = m
	}
}

func NewReconciler(mgr ctrl.Manager, opts ...ReconcilerOption) *Reconciler {
	r := &Reconciler{
		client:  mgr.GetClient(),
		scheme:  mgr.GetScheme(),
		logger:  logging.NewNopLogger(),
		record:  xpevent.NewNopRecorder(),
		metrics: metrics.NewNopRecorder(),
	}

	for _, fn := range opts {
		fn(r)
	}
	return r
}

// Reconciler reconciles a Profile object
type Reconciler struct {
	client  client.Client
	scheme  *runtime.Scheme
	logger  logging.Logger
	record  xpevent.Recorder
	metrics metrics.Recorder
}

// +kubebuilder:rbac:groups=core,resources=namespaces,verbs="*"
// +kubebuilder:rbac:groups=core,resources=serviceaccounts,verbs="*"
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs="*"
// +kubebuilder:rbac:groups=security.istio.io,resources=authorizationpolicies,verbs="*"
// +kubebuilder:rbac:groups=kubeflow.org,resources=profiles;profiles/status;profiles/finalizers,verbs="*"

// Reconcile reads that state of the cluster for a Profile object and makes changes based on the state read
// and what is in the Profile.Spec
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	profile := &v1.Profile{}
	if err := r.client.Get(ctx, req.NamespacedName, profile); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	config := &v1.ProfileConfig{}
	if err := r.client.Get(ctx, types.NamespacedName{Name: "default", Namespace: "kubeflow"}, config); err != nil {
		if !apierrors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
	}

	labels := config.Spec.Namespace.Labels

	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: profile.Name,
		},
	}
	res, err := controllerutil.CreateOrPatch(ctx, r.client, namespace, func() error {
		if err := controllerutil.SetControllerReference(profile, namespace, r.scheme); err != nil {
			return err
		}
		if namespace.Annotations == nil {
			namespace.Annotations = make(map[string]string)
		}
		namespace.Annotations["owner"] = profile.Spec.Owner.Name

		if namespace.Labels == nil {
			namespace.Labels = make(map[string]string)
		}
		for key, value := range labels {
			namespace.Labels[key] = value
		}
		return nil
	})
	if err != nil {
		return ctrl.Result{}, err
	}
	switch res {
	case controllerutil.OperationResultCreated:
		// TODO: create event
		r.metrics.CreatedNamespace(profile, namespace)
	case controllerutil.OperationResultUpdated:
		r.metrics.UpdatedNamespace(profile, namespace)
	}

	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      DefaultEditor,
			Namespace: profile.Name,
		},
	}

	res, err = controllerutil.CreateOrPatch(ctx, r.client, serviceAccount, func() error {
		return controllerutil.SetControllerReference(profile, serviceAccount, r.scheme)
	})
	if err != nil {
		return ctrl.Result{}, err
	}
	switch res {
	case controllerutil.OperationResultCreated:
		r.metrics.CreatedServiceAccount(profile, serviceAccount)
	case controllerutil.OperationResultUpdated:
		r.metrics.UpdatedServiceAccount(profile, serviceAccount)
	}

	if config.Spec.Role.Edit != "" {

		binding := &rbacv1.RoleBinding{}
		binding.SetName(serviceAccount.Name)
		binding.SetNamespace(serviceAccount.Namespace)

		res, err = controllerutil.CreateOrPatch(ctx, r.client, binding, func() error {
			if err := controllerutil.SetControllerReference(profile, binding, r.scheme); err != nil {
				return err
			}
			if binding.Annotations == nil {
				binding.Annotations = make(map[string]string)
			}
			binding.Annotations["owner"] = profile.Spec.Owner.Name
			binding.RoleRef = rbacv1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "ClusterRole",
				Name:     config.Spec.Role.Edit,
			}
			binding.Subjects = []rbacv1.Subject{{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      serviceAccount.Name,
				Namespace: profile.Name,
			}}

			return nil
		})
		if err != nil {
			return ctrl.Result{}, err
		}
		switch res {
		case controllerutil.OperationResultCreated:
			r.metrics.CreatedRoleBinding(profile, binding)
		case controllerutil.OperationResultUpdated:
			r.metrics.UpdatedRoleBinding(profile, binding)
		}
	}

	serviceAccountView := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      DefaultViewer,
			Namespace: profile.Name,
		},
	}

	res, err = controllerutil.CreateOrPatch(ctx, r.client, serviceAccountView, func() error {
		return controllerutil.SetControllerReference(profile, serviceAccountView, r.scheme)
	})
	if err != nil {
		return ctrl.Result{}, err
	}
	switch res {
	case controllerutil.OperationResultCreated:
		r.metrics.CreatedServiceAccount(profile, serviceAccountView)
	case controllerutil.OperationResultUpdated:
		r.metrics.UpdatedServiceAccount(profile, serviceAccountView)
	}

	if config.Spec.Role.View != "" {
		bindingView := &rbacv1.RoleBinding{}
		bindingView.SetName(serviceAccount.Name)
		bindingView.SetNamespace(serviceAccount.Namespace)

		res, err = controllerutil.CreateOrPatch(ctx, r.client, bindingView, func() error {
			if err := controllerutil.SetControllerReference(profile, bindingView, r.scheme); err != nil {
				return err
			}
			if bindingView.Annotations == nil {
				bindingView.Annotations = make(map[string]string)
			}
			bindingView.Annotations["owner"] = profile.Spec.Owner.Name
			bindingView.RoleRef = rbacv1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "ClusterRole",
				Name:     config.Spec.Role.View,
			}
			bindingView.Subjects = []rbacv1.Subject{{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      serviceAccount.Name,
				Namespace: profile.Name,
			}}

			return nil
		})
		if err != nil {
			return ctrl.Result{}, err
		}
		switch res {
		case controllerutil.OperationResultCreated:
			r.metrics.CreatedRoleBinding(profile, bindingView)
		case controllerutil.OperationResultUpdated:
			r.metrics.UpdatedRoleBinding(profile, bindingView)
		}

	}
	// TODO: add role for impersonate permission

	if config.Spec.Role.Admin != "" {
		ownerBinding := &rbacv1.RoleBinding{}
		ownerBinding.SetName("namespaceAdmin") // TODO: change to namespace-admin
		ownerBinding.SetNamespace(profile.Name)
		res, err = controllerutil.CreateOrPatch(ctx, r.client, ownerBinding, func() error {
			if err := controllerutil.SetControllerReference(profile, ownerBinding, r.scheme); err != nil {
				return err
			}
			if ownerBinding.Annotations == nil {
				ownerBinding.Annotations = make(map[string]string)
			}
			ownerBinding.Annotations["owner"] = profile.Spec.Owner.Name
			ownerBinding.Annotations["role"] = "admin"

			if ownerBinding.Labels == nil {
				ownerBinding.Labels = make(map[string]string)
			}
			ownerBinding.Labels[labelOwner] = profile.Name
			ownerBinding.Labels[labelRole] = "admin"

			ownerBinding.RoleRef = rbacv1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "ClusterRole",
				Name:     config.Spec.Role.Admin,
			}
			ownerBinding.Subjects = []rbacv1.Subject{profile.Spec.Owner}
			return nil
		})
		if err != nil {
			return ctrl.Result{}, err
		}
		switch res {
		case controllerutil.OperationResultCreated:
			r.metrics.CreatedRoleBinding(profile, ownerBinding)
		case controllerutil.OperationResultUpdated:
			r.metrics.UpdatedRoleBinding(profile, ownerBinding)
		}
	}

	quota := &corev1.ResourceQuota{}
	quota.SetName(ProfileQuota)
	quota.SetNamespace(profile.Name)
	res, err = controllerutil.CreateOrPatch(ctx, r.client, quota, func() error {
		if err := controllerutil.SetControllerReference(profile, quota, r.scheme); err != nil {
			return err
		}
		if len(profile.Spec.ResourceQuotaSpec.Hard) > 0 {
			quota.Spec = profile.Spec.ResourceQuotaSpec
		} else {
			if config.Spec.GlobalQuota != nil {
				quota.Spec = *config.Spec.GlobalQuota
			}
		}
		return nil
	})
	if err != nil {
		return ctrl.Result{}, err
	}
	switch res {
	case controllerutil.OperationResultCreated:
		r.metrics.CreatedResourceQuota(profile, quota)
	case controllerutil.OperationResultUpdated:
		r.metrics.UpdatedResourceQuota(profile, quota)
	}
	return ctrl.Result{}, nil
}
