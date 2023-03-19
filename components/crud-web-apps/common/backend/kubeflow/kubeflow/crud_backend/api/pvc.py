from .. import authn
from . import v1


def create_pvc(pvc, namespace, dry_run=False):
    api = v1(authn.get_username())
    return api.create_namespaced_persistent_volume_claim(
        namespace,
        pvc,
        dry_run="All" if dry_run else None,
    )


def delete_pvc(pvc, namespace):
    api = v1(authn.get_username())
    return api.delete_namespaced_persistent_volume_claim(pvc, namespace)


def list_pvcs(namespace):
    api = v1(authn.get_username())
    return api.list_namespaced_persistent_volume_claim(namespace)


def get_pvc(pvc, namespace):
    api = v1(authn.get_username())
    return api.read_namespaced_persistent_volume_claim(pvc, namespace)


def list_pvc_events(namespace, pvc_name):
    field_selector = f"involvedObject.kind=PersistentVolumeClaim,involvedObject.name={pvc_name}"
    api = v1(authn.get_username())
    return api.list_namespaced_event(
        namespace=namespace,
        field_selector=field_selector,
    )


def patch_pvc(name, namespace, pvc, auth=True):
    api = v1()
    if auth:
        api = v1(authn.get_username())
    return api.patch_namespaced_persistent_volume_claim(
        name,
        namespace,
        pvc,
    )
