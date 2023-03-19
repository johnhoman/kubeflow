from .. import authn
from . import custom_objects_api


def create_custom_rsrc(group, version, kind, data, namespace):
    api = custom_objects_api(authn.get_username())
    return api.create_namespaced_custom_object(
        group,
        version,
        namespace,
        kind,
        data,
    )


def delete_custom_rsrc(
        group,
        version,
        kind,
        name,
        namespace,
        policy="Foreground",
):
    api = custom_objects_api(authn.get_username())
    return api.delete_namespaced_custom_object(
        group,
        version,
        namespace,
        kind,
        name,
        propagation_policy=policy,
    )


def list_custom_rsrc(group, version, kind, namespace):
    api = custom_objects_api(authn.get_username())
    return api.list_namespaced_custom_object(
        group,
        version,
        namespace,
        kind,
    )


def get_custom_rsrc(group, version, kind, namespace, name):
    api = custom_objects_api(authn.get_username())
    return api.get_namespaced_custom_object(
        group,
        version,
        namespace,
        kind,
        name,
    )
