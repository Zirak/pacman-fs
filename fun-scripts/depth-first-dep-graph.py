#!/usr/bin/python2

# Experiment in depth-first traversing pkgfs, and seeing some real world usage.
# Traverses the package tree, graphing the packages and their dependencies.
# At the end, writes the graphviz notation and a jpeg of the graph to /tmp
# This is friggin cool

# Depends on networkx, pydot and graphviz.

from __future__ import print_function

import networkx
import os, os.path
import sys

graph = networkx.Graph()
repo = set()

def visit_pkg(name, path):
    if name in repo:
        return
    repo.add(name)

    print('Visting', name)
    deps_path = os.path.join(path, 'deps')

    try:
        deps = os.listdir(deps_path)
    except OSError:
        print('Fucked up dependencies:', name, file=sys.stderr)
        return

    for dep in deps:
        visit_pkg(dep, os.path.join(deps_path, dep))
        graph.add_edge(name, dep)

def main(mountpoint):
    for pkgname in os.listdir(mountpoint):
        visit_pkg(pkgname, os.path.join(mountpoint, pkgname))

if __name__ == '__main__':
    main(sys.argv[1])
    networkx.drawing.nx_pydot.write_dot(graph, '/tmp/graph')
    networkx.drawing.nx_pydot.pydot_from_networkx(graph).write_jpeg('/tmp/graph.jpeg')
