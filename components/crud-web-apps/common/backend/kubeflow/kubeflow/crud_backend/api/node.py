from . import v1


def list_nodes(api=v1()):
    return api.list_node()
