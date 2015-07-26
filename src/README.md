Hi, this is in charge of the filesystem itself. Here's what's what:

* `root.go` is the (duh) root of the filesystem. It contains some general FUSE utilities as well, but mostly it's in charge of handing out the important subdirectories.
* `index-dir.go` implements `/pkg/index`, or all of the packages available on the remote repository.
* `install-dir.go` implements `/pkg/installed`, all of the files installed locally.
* `deps-dir.go` implements a package's `deps` directory, a directory of symbolic links to the package's dependencies.
