// Hi, I'm in charge of exposing a package's dependencies as a directory.

package main

import (
	"./alpm"

	"log"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context"
)

type DepsDir struct {
	pkg *alpm.Pkg
	db  *alpm.DB
}

var _ = fs.Node(DepsDir{})

func (dir DepsDir) Attr(ctx context.Context, attr *fuse.Attr) error {
	return GenericDirAttr(ctx, attr)
}

func (dir DepsDir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	log.Println("DepsDir ReadDirAll")

	dirs := []fuse.Dirent{}

	for _, dep := range dir.pkg.GetDeps() {
		if dep.Name == "sh" {
			continue
		}

		entry := fuse.Dirent{
			Name: dep.Name,
			Type: fuse.DT_Dir,
		}

		dirs = append(dirs, entry)
	}

	return dirs, nil
}

func (dir DepsDir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	log.Println("DepsDir Lookup:", name)

	// XXX :/ why do we have to iterate *again*?
	dep := dir.db.GetProviderOf(name)

	if dep == nil {
		return nil, fuse.ENOENT
	}

	return SymlinkFile{
		// XXX sanitation
		"../../" + dep.Name(),
	}, nil
}
