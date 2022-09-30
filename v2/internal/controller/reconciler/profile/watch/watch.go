package watch

import (
	"context"
	v1 "github.com/kubeflow/kubeflow/v2/apis/core/v1"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type EnqueueRequestForProfiles struct {
	Reader client.Reader
}

func (e *EnqueueRequestForProfiles) Create(event event.CreateEvent, q workqueue.RateLimitingInterface) {
	e.add(q)
}

func (e *EnqueueRequestForProfiles) Update(event event.UpdateEvent, q workqueue.RateLimitingInterface) {
	e.add(q)
}

func (e *EnqueueRequestForProfiles) Delete(event event.DeleteEvent, q workqueue.RateLimitingInterface) {
	e.add(q)
}

func (e *EnqueueRequestForProfiles) Generic(event event.GenericEvent, q workqueue.RateLimitingInterface) {
	e.add(q)
}

func (e *EnqueueRequestForProfiles) add(q adder) {
	profileList := &v1.ProfileList{}
	if err := e.Reader.List(context.Background(), profileList); err != nil {
		return
	}

	for _, item := range profileList.Items {
		q.Add(ctrl.Request{NamespacedName: client.ObjectKeyFromObject(&item)})
	}
}

type adder interface {
	Add(interface{})
}

var _ handler.EventHandler = &EnqueueRequestForProfiles{}


func LabelPredicate(key, value string) predicate.Predicate {
	return predicate.NewPredicateFuncs(func(o client.Object) bool {
		labels := o.GetLabels()
		if labels != nil {
			return labels[key] == value
		}
		return false
	})
}