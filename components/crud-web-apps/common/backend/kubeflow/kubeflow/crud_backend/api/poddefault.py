from . import custom_objects_api
from .. import authn


def list_poddefaults(namespace):
    api = custom_objects_api(authn.get_username())
    return api.list_namespaced_custom_object(
        "kubeflow.org",
        "v1alpha1",
        namespace,
        "poddefaults",
    )
