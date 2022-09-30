package notebook

import (
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

func notebookMapper(o client.Object) []ctrl.Request {
	switch v := o.(type) {
	case *corev1.Pod:
		if v.Labels == nil {
			return nil
		}
		name, ok := v.Labels["notebook-name"]
		if !ok {
			return nil
		}
		return []ctrl.Request{{
			NamespacedName: client.ObjectKey{Name: name, Namespace: v.Namespace},
		}}
	}
	return nil
}

func podFilter(o client.Object) bool {
	pod, ok := o.(*corev1.Pod)
	if !ok {
		return false
	}
	if pod.Labels == nil {
		return false
	}
	return pod.Labels["notebook-name"] != ""
}

var (
	podPredicate = predicate.NewPredicateFuncs(podFilter)
)
