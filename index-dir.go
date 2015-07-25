package main

import (
	"./alpm"

	"log"
	"strconv"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context"
)

type IndexDir struct {
	// We have to use a slice pointer since slices aren't hashable, and bazil
	//hashes nodes. To go around that, we need to use a slice pointer. Fuck me.
	dbs *[]*alpm.DB
}

var _ = fs.Node(IndexDir{})

func (IndexDir) Attr(ctx context.Context, attr *fuse.Attr) error {
	return GenericDirAttr(ctx, attr)
}

func (dir IndexDir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	log.Println("InstallDir ReadDirAll")

	dirs := []fuse.Dirent{}

	for _, db := range *dir.dbs {
		for _, pkg := range db.GetPkgcache() {
			entry := fuse.Dirent{
				Name: pkg.Name,
				Type: fuse.DT_Dir,
			}

			dirs = append(dirs, entry)
		}
	}

	return dirs, nil
}

func (dir IndexDir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	log.Println("InstalledDir Lookup: " + name)

	var pkg *alpm.Pkg

	for _, db := range *dir.dbs {
		pkg = db.FindPackage(name)

		if pkg != nil {
			break
		}
	}

	if pkg == nil {
		return nil, fuse.ENOENT
	}

	return IndexPkgDir{pkg, dir.dbs}, nil
}

// there is a lot of code duplication between here and installed-dir. Hopefully
//we can find a way to abstract that. Together. Just you and me. That's right.
//Look into my eyes. You won't find it too forward if I hold your hand, clasp it
//between mine, treasure it? Now isn't this nice? I see you're starting to
//breathe heavily. In, out...In, out...
// Why don't we go somewhere more...comfortable?
// Like the meat grinder?

type IndexPkgDir struct {
	pkg *alpm.Pkg
	dbs *[]*alpm.DB
}

var _ = fs.Node(IndexPkgDir{})

func (dir IndexPkgDir) Attr(ctx context.Context, attr *fuse.Attr) error {
	log.Println("IndexPkgDir Attr")
	// XXX is this a good idea?
	attr.Size = uint64(dir.pkg.InstallSize)
	return GenericDirAttr(ctx, attr)
}

func (dir IndexPkgDir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	return []fuse.Dirent{
		{Name: "version", Type: fuse.DT_File},
		{Name: "description", Type: fuse.DT_File},
		{Name: "size", Type: fuse.DT_File},
		{Name: "deps", Type: fuse.DT_Dir},
	}, nil
}

func (dir IndexPkgDir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	log.Println("IndexPkgDir Lookup: " + name)

	if dir.pkg == nil {
		return nil, fuse.ENOENT
	}

	if name == "version" {
		return StupidFile{dir.pkg.Version}, nil
	}
	if name == "description" {
		return StupidFile{dir.pkg.Description}, nil
	}
	if name == "size" {
		return StupidFile{strconv.FormatInt(dir.pkg.InstallSize, 10)}, nil
	}
	if name == "deps" {
		return DepsDir{dir.pkg, dir.dbs}, nil
	}

	return nil, fuse.ENOENT
}
