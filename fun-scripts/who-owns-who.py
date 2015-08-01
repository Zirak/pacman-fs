#!/usr/bin/python2

# Experiment in traversing a package's `files` directory, to see who owns which files in your
#filesystem.
# Takes mountpoint as first argument, dumps result to path specified in second.

from __future__ import print_function

import os, os.path
import sys
import json

tree = {}

def visit_pkg(name, path):
    files_path = os.path.join(path, 'files')

    for sofar, dirs, files in os.walk(files_path):
        for f in files:
            full_path = os.path.join(sofar, f)
            linked_path = os.readlink(full_path)
            deep_set(tree, linked_path[1:].split(os.path.sep), name)

def deep_set(dictionary, key_arr, value):
    key = key_arr.pop(0)
    if key_arr:
        if key not in dictionary:
            dictionary[key] = {}
        deep_set(dictionary[key], key_arr, value)
    else:
        dictionary[key] = value

def main(mountpoint, out_path):
    for pkgname in os.listdir(mountpoint):
        print('->', pkgname)
        visit_pkg(pkgname, os.path.join(mountpoint, pkgname))

    json.dump(tree, open(out_path, 'wb'), sort_keys=True, indent=4)

if __name__ == '__main__':
    main(sys.argv[1], sys.argv[2])
