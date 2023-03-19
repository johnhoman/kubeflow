from .. import authz
from . import v1


@authz.needs_authorization("list", "core", "v1", "namespaces")
def list_namespaces(api=v1()):
    return api.list_namespace()
