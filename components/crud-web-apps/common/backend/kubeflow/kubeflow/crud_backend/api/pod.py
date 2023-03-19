from .. import authn
from . import v1


def list_pods(namespace, auth=True, label_selector = None):
    username = ""
    if auth:
        username = authn.get_username()
    api = v1(username)
    return api.list_namespaced_pod(namespace=namespace, label_selector=label_selector)


def get_pod_logs(namespace, pod, container, auth=True):
    username = ""
    if auth:
        username = authn.get_username()
    api = v1(username)
    return api.read_namespaced_pod_log(namespace=namespace, name=pod, container=container)
