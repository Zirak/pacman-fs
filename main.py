from __future__ import print_function

import fuse, fshelper
import stat, errno

root = fshelper.Directory(
    hai=fshelper.StupidFile('hai'),

    somedir=fshelper.Directory(
        file=fshelper.StupidFile('file!\n')
    )
)
anotherdir = fshelper.Directory(
    morefile=fshelper.StupidFile('more file!')
)
root.handle('anotherdir', anotherdir)

fuse.FUSE(root, 'mntdir', foreground=True)
