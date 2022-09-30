package metrics

import (
	"github.com/kubeflow/kubeflow/v2/apis/core/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

type Recorder interface {
	CreatedRoleBinding(profile *v1.Profile, role *rbacv1.RoleBinding)
	UpdatedRoleBinding(profile *v1.Profile, role *rbacv1.RoleBinding)
	DeletedRoleBinding(profile *v1.Profile, role *rbacv1.RoleBinding)
	CreatedServiceAccount(profile *v1.Profile, role *corev1.ServiceAccount)
	UpdatedServiceAccount(profile *v1.Profile, role *corev1.ServiceAccount)
	DeletedServiceAccount(profile *v1.Profile, role *corev1.ServiceAccount)
	CreatedNamespace(profile *v1.Profile, ns *corev1.Namespace)
	UpdatedNamespace(profile *v1.Profile, ns *corev1.Namespace)
	CreatedResourceQuota(profile *v1.Profile, quota *corev1.ResourceQuota)
	UpdatedResourceQuota(profile *v1.Profile, quota *corev1.ResourceQuota)
}
