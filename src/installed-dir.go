package pacmanfs

import (
	"../alpm"

	"fmt"
	"log"
	"strconv"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context"
)

type InstalledDir struct {
	db *alpm.DB
}

var _ = fs.Node(InstalledDir{})

func (InstalledDir) Attr(ctx context.Context, attr *fuse.Attr) error {
	log.Println("InstalledDir Attr")
	return GenericDirAttr(ctx, attr)
}

func (dir InstalledDir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	log.Println("InstallDir ReadDirAll")

	dirs := []fuse.Dirent{}

	for _, pkg := range dir.db.GetPkgcache() {
		entry := fuse.Dirent{
			Name: pkg.Name,
			Type: fuse.DT_Dir,
		}

		dirs = append(dirs, entry)
	}

	return dirs, nil
}

func (dir InstalledDir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	log.Println("InstalledDir Lookup: " + name)

	pkg := dir.db.FindPackage(name)

	if pkg == nil {
		log.Println("InstalledDir Not found:", name)
		return nil, fuse.ENOENT
	}

	return InstalledPkgDir{pkg, dir.db}, nil
}

type InstalledPkgDir struct {
	pkg *alpm.Pkg
	db  *alpm.DB
}

var _ = fs.Node(InstalledPkgDir{})

func (dir InstalledPkgDir) Attr(ctx context.Context, attr *fuse.Attr) error {
	log.Println("InstalledPkgDir Attr")
	// XXX is this a good idea?
	attr.Size = uint64(dir.pkg.InstallSize)
	return GenericDirAttr(ctx, attr)
}

func (dir InstalledPkgDir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	return []fuse.Dirent{
		{Name: "version", Type: fuse.DT_File},
		{Name: "description", Type: fuse.DT_File},
		{Name: "size", Type: fuse.DT_File},

		{Name: "dependencies", Type: fuse.DT_Dir},
		{Name: "files", Type: fuse.DT_Dir},

		{Name: "uninstall", Type: fuse.DT_File},
	}, nil
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
		return StupidFile{Contents: dir.pkg.Version}, nil
	}
	if name == "description" {
		return StupidFile{Contents: dir.pkg.Description}, nil
	}
	if name == "size" {
		return StupidFile{Contents: strconv.FormatInt(dir.pkg.InstallSize, 10)}, nil
	}

	if name == "dependencies" {
		return DepsDir{dir.pkg, &[]*alpm.DB{dir.db}}, nil
	}
	if name == "files" {
		return filesToStupidDir(dir.pkg.GetFiles()), nil
	}

	if name == "uninstall" {
		return StupidFile{
			Contents: fmt.Sprintf(`#!/bin/sh
pacman -R '%s'
`, dir.pkg.Name),
			Mode: 0555,
		}, nil
	}

	return nil, fuse.ENOENT
}
