package main

import (
	"./alpm"

	"log"
	"strconv"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context"
)

type InstalledDir struct {
	handle *alpm.Handle
	db     *alpm.DB
}

var _ = fs.Node(InstalledDir{})

func (InstalledDir) Attr(ctx context.Context, attr *fuse.Attr) error {
	log.Println("InstalledDir Attr")
	return GenericDirAttr(ctx, attr)
}

func (dir InstalledDir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	log.Println("InstallDir ReadDirAll")

	dirs := []fuse.Dirent{}

	for _, pkg := range dir.db.GetPkgcache().Slice() {
		entry := fuse.Dirent{
			Name: pkg.Name(),
			Type: fuse.DT_Dir,
		}

		dirs = append(dirs, entry)
	}

	return dirs, nil
}

func (dir InstalledDir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	log.Println("InstalledDir Lookup: " + name)

	var pkg alpm.Pkg
	// XXX
	found := false

	// XXX :/
	for _, p := range dir.db.GetPkgcache().Slice() {
		if p.Name() == name {
			pkg = p
			found = true
			break
		}
	}

	if !found {
		log.Println("Installed: Not found")
		return nil, fuse.ENOENT
	}

	return InstalledPkgDir{&pkg, dir.db}, nil
}

type InstalledPkgDir struct {
	pkg *alpm.Pkg
	db  *alpm.DB
}

var pkgDirEntries = []fuse.Dirent{
	{Name: "version", Type: fuse.DT_File},
	{Name: "desc", Type: fuse.DT_File},
	{Name: "size", Type: fuse.DT_File},
	{Name: "deps", Type: fuse.DT_Dir},
}

var _ = fs.Node(InstalledPkgDir{})

func (dir InstalledPkgDir) Attr(ctx context.Context, attr *fuse.Attr) error {
	log.Println("InstalledPkgDir Attr")
	// XXX is this a good idea?
	attr.Size = uint64(dir.pkg.InstallSize())
	return GenericDirAttr(ctx, attr)
}

func (dir InstalledPkgDir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	return pkgDirEntries, nil
}

func (dir InstalledPkgDir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	log.Println("InstalledPkgDir Lookup: " + name)

	// XXX really weird bug: when `ls`ing with fish, we get a lookup for
	//.gitignore. InstallDir.Lookup returns ENOENT, but this function is still
	//called trying to look it up, with dir.pkg being the first package.
	// I tried writing the following check to get around that, but no dice. I
	//dunno what else to try.
	if dir.pkg == nil {
		log.Println("Not found bitches")
		return nil, fuse.ENOENT
	}

	if name == "version" {
		return StupidFile{dir.pkg.Version()}, nil
	}
	if name == "desc" {
		return StupidFile{dir.pkg.Desc()}, nil
	}
	if name == "size" {
		return StupidFile{strconv.FormatInt(dir.pkg.InstallSize(), 10)}, nil
	}
	if name == "deps" {
		return DepsDir{dir.pkg, dir.db}, nil
	}

	return nil, fuse.ENOENT
}
