package metrics

import (
	v1 "github.com/kubeflow/kubeflow/v2/apis/core/v1"
	"github.com/prometheus/client_golang/prometheus"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	counterCreatedRoleBinding = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "kubeflow",
			Subsystem: "profile_controller",
			Name:      "created_role_binding",
		},
		[]string{"profile", "role", "serviceaccount", "owner"},
	)
	counterUpdatedRoleBinding = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "kubeflow",
			Subsystem: "profile_controller",
			Name:      "updated_role_binding",
		},
		[]string{"profile", "role", "serviceaccount", "owner"},
	)
	counterDeletedRoleBinding = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "kubeflow",
			Subsystem: "profile_controller",
			Name:      "deleted_role_binding",
		},
		[]string{"profile", "role", "serviceaccount", "owner"},
	)
)

func init() {
	metrics.Registry.MustRegister(
		counterCreatedRoleBinding,
		counterUpdatedRoleBinding,
		counterDeletedRoleBinding,
	)
}

type Prometheus struct{}

func (p *Prometheus) CreatedRoleBinding(profile *v1.Profile, role *rbacv1.RoleBinding) {
	counterCreatedRoleBinding.WithLabelValues(profile.Name, role.Name, role.Subjects[0].Name, profile.Spec.Owner.Name).Inc()
}
func (p *Prometheus) UpdatedRoleBinding(profile *v1.Profile, role *rbacv1.RoleBinding) {
	counterUpdatedRoleBinding.WithLabelValues(profile.Name, role.Name, role.Subjects[0].Name, profile.Spec.Owner.Name).Inc()
}
func (p *Prometheus) DeletedRoleBinding(profile *v1.Profile, role *rbacv1.RoleBinding) {
	counterDeletedRoleBinding.WithLabelValues(profile.Name, role.Name, role.Subjects[0].Name, profile.Spec.Owner.Name).Inc()
}
func (p *Prometheus) CreatedServiceAccount(profile *v1.Profile, role *corev1.ServiceAccount) {}
func (p *Prometheus) UpdatedServiceAccount(profile *v1.Profile, role *corev1.ServiceAccount) {}
func (p *Prometheus) DeletedServiceAccount(profile *v1.Profile, role *corev1.ServiceAccount) {}
func (p *Prometheus) CreatedNamespace(profile *v1.Profile, ns *corev1.Namespace)             {}
func (p *Prometheus) UpdatedNamespace(profile *v1.Profile, ns *corev1.Namespace)             {}
func (p *Prometheus) CreatedResourceQuota(profile *v1.Profile, quota *corev1.ResourceQuota)  {}
func (p *Prometheus) UpdatedResourceQuota(profile *v1.Profile, quota *corev1.ResourceQuota)  {}

func NewPrometheusRecorder() *Prometheus { return &Prometheus{} }

var _ Recorder = &Prometheus{}
