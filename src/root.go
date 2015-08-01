package pacmanfs

import (
	"../alpm"

	"log"
	"os"

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
