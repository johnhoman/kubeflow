package configmap

import (
	"context"
	"encoding/json"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sort"

	xpevent "github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/kubeflow/kubeflow/v2/apis/centraldashboard/v1alpha1"
	"github.com/kubeflow/kubeflow/v2/internal/logging"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	reasonCreatedConfigMap xpevent.Reason = "CreatedConfigMap"
	reasonUpdatedConfigMap xpevent.Reason = "UpdatedConfigMap"
)

const (
	configmapName = "centraldashboard-config"
	fieldLinks    = "links"
)

func Setup(mgr ctrl.Manager, o controller.Options) error {

	const name = "kubeflow/reflect/configmap"

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(
			&corev1.ConfigMap{},
			builder.WithPredicates(predicate.NewPredicateFuncs(func(o client.Object) bool {
				return o.GetName() == configmapName && o.GetNamespace() == "kubeflow"
			})),
		).
		Watches(
			&source.Kind{Type: &v1alpha1.MenuLink{}},
			handler.EnqueueRequestsFromMapFunc(func(_ client.Object) []ctrl.Request {
				return []ctrl.Request{{NamespacedName: client.ObjectKey{Name: configmapName, Namespace: "kubeflow"}}}
			}),
		).
		WithOptions(o).
		Complete(NewReconciler(mgr,
			WithLogger(logging.NewLogrLogger(mgr.GetLogger().WithValues("controller", name))),
			WithEventRecorder(xpevent.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		))
}

type ReconcilerOption func(r *Reconciler)

func WithLogger(log logging.Logger) ReconcilerOption {
	return func(r *Reconciler) {
		r.logger = log
	}
}

func WithEventRecorder(recorder xpevent.Recorder) ReconcilerOption {
	return func(r *Reconciler) {
		r.record = recorder
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

type Reconciler struct {
	client client.Client
	scheme *runtime.Scheme
	logger logging.Logger
	record xpevent.Recorder
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	configMap := &corev1.ConfigMap{}
	if err := r.client.Get(ctx, req.NamespacedName, configMap); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	linkConfig := &Links{}

	if configMap.Data == nil {
		configMap.Data = make(map[string]string)
	}

	if links := configMap.Data[fieldLinks]; links != "" {
		if err := json.Unmarshal([]byte(links), linkConfig); err != nil {
			return ctrl.Result{}, err
		}
	}

	if linkConfig.MenuLinks == nil {
		linkConfig.MenuLinks = make([]v1alpha1.MenuLinkSpec, 0)
	}
	if linkConfig.QuickLinks == nil {
		linkConfig.QuickLinks = make([]v1alpha1.QuickLinkSpec, 0)
	}
	if linkConfig.DocumentationItems == nil {
		linkConfig.DocumentationItems = make([]v1alpha1.DocumentationItemSpec, 0)
	}

	menuList := &v1alpha1.MenuLinkList{}
	if err := r.client.List(ctx, menuList); err != nil {
		return ctrl.Result{}, err
	}
	for _, item := range menuList.Items {
		linkConfig.MenuLinks = append(linkConfig.MenuLinks, item.Spec)
	}
	sort.Slice(linkConfig.MenuLinks, func(i, j int) bool {
		return linkConfig.MenuLinks[i].Link < linkConfig.MenuLinks[j].Link
	})

	quickLinkList := &v1alpha1.QuickLinkList{}
	if err := r.client.List(ctx, quickLinkList); err != nil {
		return ctrl.Result{}, err
	}
	for _, item := range quickLinkList.Items {
		linkConfig.QuickLinks = append(linkConfig.QuickLinks, item.Spec)
	}
	sort.Slice(linkConfig.QuickLinks, func(i, j int) bool {
		return linkConfig.QuickLinks[i].Link < linkConfig.QuickLinks[j].Link
	})

	documentationItemList := &v1alpha1.DocumentationItemList{}
	if err := r.client.List(ctx, documentationItemList); err != nil {
		return ctrl.Result{}, err
	}
	for _, item := range documentationItemList.Items {
		linkConfig.DocumentationItems = append(linkConfig.DocumentationItems, item.Spec)
	}
	sort.Slice(linkConfig.DocumentationItems, func(i, j int) bool {
		return linkConfig.DocumentationItems[i].Link < linkConfig.DocumentationItems[j].Link
	})

	raw, err := json.Marshal(linkConfig)
	if err != nil {
		return ctrl.Result{}, err
	}

	patch := client.MergeFrom(configMap.DeepCopy())

	configMap.Data["settings"] = `{"DASHBOARD_FORCE_IFRAME": true}`
	configMap.Data[fieldLinks] = string(raw)
	if err := r.client.Patch(ctx, configMap, patch); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}
