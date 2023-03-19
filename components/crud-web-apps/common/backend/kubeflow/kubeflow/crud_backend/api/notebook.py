from .. import authn
from . import custom_objects_api, v1


def get_notebook(notebook, namespace):
    api = custom_objects_api(authn.get_username())
    return api.get_namespaced_custom_object(
        "kubeflow.org", "v1beta1", namespace, "notebooks", notebook
    )


def create_notebook(notebook, namespace, dry_run=False):
    api = custom_objects_api(authn.get_username())
    return api.create_namespaced_custom_object(
        "kubeflow.org", "v1beta1", namespace, "notebooks", notebook,
        dry_run="All" if dry_run else None)


def list_notebooks(namespace):
    api = custom_objects_api(authn.get_username())
    return api.list_namespaced_custom_object("kubeflow.org", "v1beta1", namespace, "notebooks")


def delete_notebook(notebook, namespace):
    api = custom_objects_api(authn.get_username())
    return api.delete_namespaced_custom_object(
        group="kubeflow.org",
        version="v1beta1",
        namespace=namespace,
        plural="notebooks",
        name=notebook,
        propagation_policy="Foreground",
    )


def patch_notebook(notebook, namespace, body):
    api = custom_objects_api(authn.get_username())
    return api.patch_namespaced_custom_object(
        "kubeflow.org", "v1beta1", namespace, "notebooks", notebook, body
    )


def list_notebook_events(notebook, namespace):
    selector = "involvedObject.kind=Notebook,involvedObject.name=" + notebook
    api = v1(authn.get_username())
    return api.list_namespaced_event(namespace=namespace, field_selector=selector)
