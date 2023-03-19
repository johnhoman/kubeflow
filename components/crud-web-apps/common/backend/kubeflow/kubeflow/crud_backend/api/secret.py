from .. import authn
from . import v1


def get_secret(namespace, name, auth=True, api=v1()):
    if auth:
        api = v1(authn.get_username())
    return api.read_namespaced_secret(name, namespace)


def create_secret(namespace, secret, auth=True, api=v1()):
    if auth:
        api = v1(authn.get_username())
    return api.create_namespaced_secret(namespace, secret)
