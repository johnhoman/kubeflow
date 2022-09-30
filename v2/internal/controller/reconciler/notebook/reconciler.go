/*

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

package notebook

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	xpevent "github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/kubeflow/kubeflow/v2/apis/core/v1"
	"github.com/kubeflow/kubeflow/v2/internal/controller/reconciler/notebook/culler"
	"github.com/kubeflow/kubeflow/v2/internal/logging"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	DefaultContainerPort        = 8888
	DefaultServingPort          = 80
	AnnotationRewriteURI        = "notebooks.kubeflow.org/http-rewrite-uri"
	AnnotationHeadersRequestSet = "notebooks.kubeflow.org/http-headers-request-set"
	PrefixEnvVar                = "NB_PREFIX"
	// DefaultFSGroup is the default fsGroup of PodSecurityContext.
	// https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.11/#podsecuritycontext-v1-core
	DefaultFSGroup = int64(100)
)

const (
	reasonStatefulSetUpdateFailed = "StatefulSetUpdateFailed"
)

func Setup(mgr ctrl.Manager, o controller.Options) error {

	name := "kubeflow/" + strings.ToLower(v1.NotebookKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1.Notebook{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.Service{}).
		Watches(
			&source.Kind{Type: &corev1.Pod{}},
			handler.EnqueueRequestsFromMapFunc(notebookMapper),
			builder.WithPredicates(podPredicate),
		).
		Complete(NewReconciler(mgr,
			WithLogger(logging.NewLogrLogger(mgr.GetLogger().WithValues("controller", name))),
			WithEventRecorder(xpevent.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			WithFSGroup(os.Getenv("ADD_FSGROUP")),
		))
}

type ReconcilerOption func(r *Reconciler)

func WithLogger(logger logging.Logger) ReconcilerOption {
	return func(r *Reconciler) {
		r.logger = logger
	}
}

func WithEventRecorder(reason xpevent.Recorder) ReconcilerOption {
	return func(r *Reconciler) {
		r.record = reason
	}
}

func WithFSGroup(grp string) ReconcilerOption {
	return func(r *Reconciler) {
		if grp != "" {
			r.fsGroup = pointer.String(grp)
		}
	}
}

func NewReconciler(mgr ctrl.Manager, opts ...ReconcilerOption) *Reconciler {
	r := &Reconciler{
		client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
		logger: logging.NewNopLogger(),
		record: xpevent.NewNopRecorder(),
	}

	for _, fn := range opts {
		fn(r)
	}
	return r
}

// Reconciler reconciles a Notebook object
type Reconciler struct {
	client client.Client
	scheme *runtime.Scheme

	logger  logging.Logger
	record  xpevent.Recorder
	fsGroup *string
}

// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create;patch
// +kubebuilder:rbac:groups=core,resources=services,verbs="*"
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs="*"
// +kubebuilder:rbac:groups=kubeflow.org,resources=notebooks;notebooks/status;notebooks/finalizers,verbs="*"
// +kubebuilder:rbac:groups="networking.istio.io",resources=virtualservices,verbs="*"

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// If not found, continue. Is not an event.

	notebook := &v1.Notebook{}
	if err := r.client.Get(ctx, req.NamespacedName, notebook); err != nil {
		r.logger.Debug(err.Error(), "unable to fetch Notebook")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// jupyter-web-app deletes objects using foreground deletion policy, Notebook CR will stay until all owned objects are deleted
	// reconcile loop might keep on trying to recreate the resources that the API server tries to delete.
	// so when Notebook CR is terminating, reconcile loop should do nothing
	if !notebook.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      notebook.Name,
			Namespace: notebook.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: pointer.Int32(1),
		},
	}

	// Reconcile StatefulSet
	res, err := controllerutil.CreateOrPatch(ctx, r.client, sts, func() error {
		if err := controllerutil.SetControllerReference(notebook, sts, r.scheme); err != nil {
			return err
		}

		if sts.Labels == nil {
			sts.Labels = make(map[string]string)
		}
		if sts.Annotations == nil {
			sts.Annotations = make(map[string]string)
		}

		if sts.Annotations[culler.AnnotationStop] != "true" {
			sts.Spec.Replicas = pointer.Int32(1)
		}

		// Copy all the labels
		for key, value := range notebook.Labels {
			sts.Spec.Template.ObjectMeta.Labels[key] = value
		}

		// don't touch the replicas after creation. The culler
		// should manage that
		// sts.Spec.Replicas = pointer.Int32(1)

		sts.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"statefulset": notebook.Name,
			},
		}
		sts.Spec.Template = corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{
				"statefulset":   notebook.Name,
				"notebook-name": notebook.Name,
			}},
			Spec: notebook.Spec.Template.Spec,
		}

		podSpec := &sts.Spec.Template.Spec
		// TODO: remove this from environment variable
		if value, exists := os.LookupEnv("ADD_FSGROUP"); !exists || value == "true" {
			if podSpec.SecurityContext == nil {
				fsGroup := DefaultFSGroup
				podSpec.SecurityContext = &corev1.PodSecurityContext{
					FSGroup: &fsGroup,
				}
			}
		}

		container := &sts.Spec.Template.Spec.Containers[0]
		if container.WorkingDir == "" {
			container.WorkingDir = "/home/jovyan"
		}
		if container.Ports == nil {
			container.Ports = []corev1.ContainerPort{{
				ContainerPort: DefaultContainerPort,
				Name:          "notebook-port",
				Protocol:      corev1.ProtocolTCP,
			}}
		}

		prefix := fmt.Sprintf("/notebook/%s/%s", sts.Namespace, sts.Name)
		vars := make(map[string]int)
		for k, env := range container.Env {
			vars[env.Name] = k
		}
		env := corev1.EnvVar{Name: PrefixEnvVar, Value: prefix}
		if k, ok := vars[PrefixEnvVar]; ok {
			container.Env[k] = env
		} else {
			container.Env = append(container.Env, env)
		}

		return nil
	})

	if err != nil {
		r.record.Event(notebook, xpevent.Warning(reasonStatefulSetUpdateFailed, err))
		return ctrl.Result{}, err
	}
	switch res {
	case controllerutil.OperationResultCreated:
	case controllerutil.OperationResultUpdated:
	}

	// Reconcile service
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      notebook.Name,
			Namespace: notebook.Namespace,
		},
	}
	res, err = controllerutil.CreateOrPatch(ctx, r.client, service, func() error {
		port := DefaultContainerPort
		containerPorts := notebook.Spec.Template.Spec.Containers[0].Ports
		if containerPorts != nil {
			port = int(containerPorts[0].ContainerPort)
		}
		service.Spec = corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: map[string]string{"statefulset": notebook.Name},
			Ports: []corev1.ServicePort{{
				// Make port name follow Istio pattern so it can be managed by istio rbac
				Name:       "http-" + notebook.Name,
				Port:       DefaultServingPort,
				TargetPort: intstr.FromInt(port),
				Protocol:   corev1.ProtocolTCP,
			},
			},
		}
		return nil
	})

	if err != nil {
		r.record.Event(notebook, xpevent.Warning(xpevent.Reason("TODO"), err))
		return ctrl.Result{}, err
	}
	switch res {
	case controllerutil.OperationResultCreated:
	case controllerutil.OperationResultUpdated:
	}

	pod := &corev1.Pod{}
	if err := r.client.Get(ctx, client.ObjectKey{Name: sts.Name + "-0", Namespace: sts.Namespace}, pod); err != nil {
		if !apierrs.IsNotFound(err) {
			return ctrl.Result{}, err
		}
	} else {
		err := updateNotebookStatus(r, notebook, sts, pod, req)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	// Update Notebook CR status
	err = updateNotebookStatus(r, notebook, sts, pod, req)
	if err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func updateNotebookStatus(r *Reconciler, nb *v1.Notebook,
	sts *appsv1.StatefulSet, pod *corev1.Pod, req ctrl.Request) error {

	log := r.logger.WithValues("notebook", req.NamespacedName)
	ctx := context.Background()

	status, err := createNotebookStatus(r, nb, sts, pod, req)
	if err != nil {
		return err
	}

	log.Info("Updating Notebook CR Status", "status", status)
	patch := client.MergeFrom(nb)
	nb.Status = status
	return r.client.Status().Patch(ctx, nb, patch)
}

func createNotebookStatus(r *Reconciler, nb *v1.Notebook,
	sts *appsv1.StatefulSet, pod *corev1.Pod, req ctrl.Request) (v1.NotebookStatus, error) {

	// Initialize Notebook CR Status
	r.logger.Info("Initializing Notebook CR Status")
	status := v1.NotebookStatus{
		Conditions:     make([]v1.NotebookCondition, 0),
		ReadyReplicas:  sts.Status.ReadyReplicas,
		ContainerState: corev1.ContainerState{},
	}

	// Update the status based on the Pod's status
	if reflect.DeepEqual(pod.Status, corev1.PodStatus{}) {
		r.logger.Info("No pod.Status found. Won't update notebook conditions and containerState")
		return status, nil
	}

	// Update status of the CR using the ContainerState of
	// the container that has the same name as the CR.
	// If no container of same name is found, the state of the CR is not updated.
	notebookContainerFound := false
	r.logger.Info("Calculating Notebook's  containerState")
	for i := range pod.Status.ContainerStatuses {
		if pod.Status.ContainerStatuses[i].Name != nb.Name {
			continue
		}

		if pod.Status.ContainerStatuses[i].State == nb.Status.ContainerState {
			continue
		}

		// Update Notebook CR's status.ContainerState
		cs := pod.Status.ContainerStatuses[i].State
		r.logger.Info("Updating Notebook CR state: ", "state", cs)

		status.ContainerState = cs
		notebookContainerFound = true
		break
	}

	if !notebookContainerFound {
		r.logger.Debug(
			"Could not find container with the same name as Notebook " +
				"in containerStates of Pod. Will not update notebook's " +
				"status.containerState ",
		)
	}

	// Mirroring pod condition
	notebookConditions := make([]v1.NotebookCondition, 0)
	r.logger.Info("Calculating Notebook's Conditions")
	for i := range pod.Status.Conditions {
		condition := PodCondToNotebookCond(pod.Status.Conditions[i])
		notebookConditions = append(notebookConditions, condition)
	}

	status.Conditions = notebookConditions

	return status, nil
}

func PodCondToNotebookCond(podc corev1.PodCondition) v1.NotebookCondition {

	condition := v1.NotebookCondition{}

	if len(podc.Type) > 0 {
		condition.Type = string(podc.Type)
	}

	if len(podc.Status) > 0 {
		condition.Status = string(podc.Status)
	}

	if len(podc.Message) > 0 {
		condition.Message = podc.Message
	}

	if len(podc.Reason) > 0 {
		condition.Reason = podc.Reason
	}

	// check if podc.LastProbeTime is null. If so initialize
	// the field with metav1.Now()
	check := podc.LastProbeTime.Time.Equal(time.Time{})
	if !check {
		condition.LastProbeTime = podc.LastProbeTime
	} else {
		condition.LastProbeTime = metav1.Now()
	}

	// check if podc.LastTransitionTime is null. If so initialize
	// the field with metav1.Now()
	check = podc.LastTransitionTime.Time.Equal(time.Time{})
	if !check {
		condition.LastTransitionTime = podc.LastTransitionTime
	} else {
		condition.LastTransitionTime = metav1.Now()
	}

	return condition
}
