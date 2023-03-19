from functools import lru_cache

from kubernetes import client, config
from kubernetes.config import ConfigException

try:
    config.load_incluster_config()
except ConfigException:
    config.load_kube_config()


def impersonate(username: str = ""):
    if username:
        return client.ApiClient(header_name="Impersonate-User", header_value=username)
    return client.ApiClient()


@lru_cache(maxsize=100)
def v1(username: str = ""):
    return client.CoreV1Api(impersonate(username))


@lru_cache(maxsize=100)
def storage_v1(username: str = ""):
    return client.StorageV1Api(impersonate(username))


@lru_cache(maxsize=100)
def custom_objects_api(username: str = ""):
    return client.CustomObjectsApi(impersonate(username))