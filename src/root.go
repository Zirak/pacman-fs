package pacmanfs

import (
	"../alpm"

	"log"
	"os"
	"strings"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context"
)

type DBWrapper struct {
	local *alpm.DB
	// an explanation on why we use a slice pointer instead of a slice is
	//available on Amazon for $9.99, limited Christmas edition all year round.
	// It's also available in index-dir.go
	sync *[]*alpm.DB
}

type FS struct {
	dbs *DBWrapper
}

var _ fs.FS = (*FS)(nil)

func NewFilesystem(localdb *alpm.DB, syncdbs *[]*alpm.DB) *FS {
	wrapper := &DBWrapper{local: localdb, sync: syncdbs}
	return &FS{wrapper}
}

func (filesys FS) Root() (fs.Node, error) {
	root := &RootDir{filesys.dbs}
	return root, nil
}

type RootDir struct {
	*DBWrapper
}

var _ = fs.Node(RootDir{})

func (RootDir) Attr(ctx context.Context, attr *fuse.Attr) error {
	return GenericDirAttr(ctx, attr)
}

func (RootDir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	return []fuse.Dirent{
		{Name: "installed", Type: fuse.DT_Dir},
		{Name: "index", Type: fuse.DT_Dir},

		{Name: "sync", Type: fuse.DT_File},
	}, nil
}

func (dir RootDir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	if name == "installed" {
		return InstalledDir{dir.local}, nil
	}

	if name == "index" {
		return IndexDir{dir.sync}, nil
	}

	if name == "sync" {
		// XXX test whether we need to re-get the sync dbs after this is run
		return NewExecutableFile(`#!/bin/sh
pacman -Sy
`), nil
	}

	// sshhh, don't tell anyone
	if name == "pacman" {
		return NewStaticFile(` .--.
/ _.-' .-.  .-.  .-.
\  '-. '-'  '-'  '-'
 '--'`), nil
	}

	return nil, fuse.ENOENT
}

// And now! A few helpers.

func GenericDirAttr(ctx context.Context, attr *fuse.Attr) error {
	attr.Mode = os.ModeDir | 0555
	return nil
}

type SymlinkFile struct {
	target string
}

var _ = fs.Node(&SymlinkFile{})
var _ = fs.NodeReadlinker(&SymlinkFile{})

func (file SymlinkFile) Attr(ctx context.Context, attr *fuse.Attr) error {
	attr.Mode = os.ModeSymlink
	attr.Size = uint64(len(file.target))
	return nil
}

func (file SymlinkFile) Readlink(ctx context.Context, req *fuse.ReadlinkRequest) (string, error) {
	log.Println("SymlinkFile Readlink to: " + file.target)
	return file.target, nil
}

type StupidDir struct {
	// I hate this, but I can't figure out a better alternative.
	// I wanted to only have one map, Children, which themselves either have a
	//Children property signifying a StupidDir, or a Content property
	//signifying a StupidFile.
	Files *map[string]fs.Node
	Dirs  *map[string]*StupidDir
}

var _ = fs.Node(&StupidDir{})
var _ = fs.HandleReadDirAller(&StupidDir{})

func NewStupidDir() *StupidDir {
	files := make(map[string]fs.Node)
	dirs := make(map[string]*StupidDir)

	return &StupidDir{
		Files: &files,
		Dirs:  &dirs,
	}
}

func filesToStupidDir(files []*alpm.File) *StupidDir {
	// We need to parse this slice of path strings:
	/*
		etc/
		etc/zsh/
		etc/zsh/zprofile
		etc/moose
	*/
	// Into an actual StupidDir structure, so the above will turn into:
	/*
		StupidDir{
			Files: {},
			Dirs: {
				"etc": StupidDir{
					Files: { "moose": StupidFile{} },
					// you get the point
				}
			}
		}
	*/
	ret := NewStupidDir()

	// I hate myself so much
	// That this is not a haiku
	// Well, it's almost one

	for _, file := range files {
		path := file.Name
		isDir := len(path)-1 == strings.LastIndex(path, "/")
		// remove trailing /
		if isDir {
			path = path[:len(path)-1]
		}

		// "/etc/zsh/zprofile" -> [etc zsh zprofile]
		parts := strings.Split(path, "/")
		name := parts[len(parts)-1]
		leading := parts[:len(parts)-1]

		// traverse the tree until we get to the parent
		parent := ret
		for _, dirname := range leading {
			parent, _ = (*parent.Dirs)[dirname]
		}

		// add ourselves to the tree
		if isDir {
			(*parent.Dirs)[name] = NewStupidDir()
		} else {
			(*parent.Files)[name] = SymlinkFile{"/" + path}
		}
	}

	return ret
	// My eyes are bleeding
	// There is no hope for mankind
	// Doughnuts are tasty
}

func (StupidDir) Attr(ctx context.Context, attr *fuse.Attr) error {
	return GenericDirAttr(ctx, attr)
}

func (dir StupidDir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	ret := []fuse.Dirent{}

	for filename, _ := range *dir.Files {
		ret = append(ret, fuse.Dirent{Name: filename, Type: fuse.DT_File})
	}
	for dirname, _ := range *dir.Dirs {
		ret = append(ret, fuse.Dirent{Name: dirname, Type: fuse.DT_Dir})
	}

	return ret, nil
}

func (dir StupidDir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	file, ok := (*dir.Files)[name]
	if ok {
		return file, nil
	}

	child, ok := (*dir.Dirs)[name]
	if ok {
		return *child, nil
	}

	return nil, fuse.ENOENT
}

type StupidFile struct {
	Contents string
	Mode     os.FileMode
}

var _ = fs.Node(&StupidFile{})
var _ = fs.HandleReadAller(&StupidFile{})

func NewStaticFile(contents string) *StupidFile {
	return &StupidFile{
		Contents: contents,
		Mode:     0444,
	}
}

func NewExecutableFile(contents string) *StupidFile {
	return &StupidFile{
		Contents: contents,
		Mode:     0555,
	}
}

func (file StupidFile) Attr(ctx context.Context, attr *fuse.Attr) error {
	log.Println("StupidFile Attr")

	// XXX figure out some Inode scheme.

	if file.Mode != 0 {
		attr.Mode = file.Mode
	} else {
		attr.Mode = 0444
	}

	attr.Size = uint64(len(file.Contents))
	return nil
}

func (file StupidFile) ReadAll(ctx context.Context) ([]byte, error) {
	log.Println("File ReadAll")
	return []byte(file.Contents), nil
}
