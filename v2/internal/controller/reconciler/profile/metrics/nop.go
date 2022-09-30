package metrics

import (
	"github.com/kubeflow/kubeflow/v2/apis/core/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

func NewNopRecorder() *nopRecorder { return &nopRecorder{} }

type nopRecorder struct{}

func (n *nopRecorder) CreatedServiceAccount(*v1.Profile, *corev1.ServiceAccount) {}
func (n *nopRecorder) UpdatedServiceAccount(*v1.Profile, *corev1.ServiceAccount) {}
func (n *nopRecorder) DeletedServiceAccount(*v1.Profile, *corev1.ServiceAccount) {}
func (n *nopRecorder) CreatedResourceQuota(*v1.Profile, *corev1.ResourceQuota)   {}
func (n *nopRecorder) UpdatedResourceQuota(*v1.Profile, *corev1.ResourceQuota)   {}
func (n *nopRecorder) CreatedRoleBinding(*v1.Profile, *rbacv1.RoleBinding)       {}
func (n *nopRecorder) UpdatedRoleBinding(*v1.Profile, *rbacv1.RoleBinding)       {}
func (n *nopRecorder) DeletedRoleBinding(*v1.Profile, *rbacv1.RoleBinding)       {}
func (n *nopRecorder) CreatedNamespace(*v1.Profile, *corev1.Namespace)           {}
func (n *nopRecorder) UpdatedNamespace(*v1.Profile, *corev1.Namespace)           {}

var _ Recorder = &nopRecorder{}
