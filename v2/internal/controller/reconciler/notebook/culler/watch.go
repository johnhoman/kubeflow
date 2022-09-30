package culler

import (
	"context"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	corev1beta1 "github.com/kubeflow/kubeflow/v2/apis/core/v1"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
)

type EnqueueRequestForNotebooks struct {
	reader client.Reader
}

func (e *EnqueueRequestForNotebooks) Create(event event.CreateEvent, q workqueue.RateLimitingInterface) {
	e.add(event.Object, q)
}

func (e *EnqueueRequestForNotebooks) Update(event event.UpdateEvent, q workqueue.RateLimitingInterface) {
	e.add(event.ObjectOld, q)
	e.add(event.ObjectNew, q)
}

func (e *EnqueueRequestForNotebooks) Delete(event event.DeleteEvent, q workqueue.RateLimitingInterface) {
}

func (e *EnqueueRequestForNotebooks) Generic(event event.GenericEvent, q workqueue.RateLimitingInterface) {
	e.add(event.Object, q)
}

func (e *EnqueueRequestForNotebooks) add(o client.Object, q workqueue.RateLimitingInterface) {
	opts := make([]client.ListOption, 0)
	if o.GetNamespace() != "kubeflow" {
		opts = append(opts, client.InNamespace(o.GetNamespace()))
	}

	notebookList := &corev1beta1.NotebookList{}
	if err := e.reader.List(context.Background(), notebookList, opts...); err != nil {
		return
	}

	for _, item := range notebookList.Items {
		q.Add(ctrl.Request{NamespacedName: client.ObjectKeyFromObject(&item)})
	}
}

var _ handler.EventHandler = &EnqueueRequestForNotebooks{}

func stsFilter(o client.Object) bool {
	if o.GetLabels() != nil && o.GetLabels()["notebook-name"] != "" {
		return true
	}
	return false
}

var stsPredicate = predicate.NewPredicateFuncs(stsFilter)
