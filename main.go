package main

import (
	"./alpm"

	"flag"
	"fmt"
	"log"
	"os"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context"
)

func main() {
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "Pass a mountpoint, dummy.")
		os.Exit(1)
	}

	handle, err := alpm.Init()
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Release()

	db, err := handle.GetLocalDb()
	if err != nil {
		log.Fatal(err)
	}

	mountpoint := flag.Arg(0)

	client, err := fuse.Mount(
		mountpoint,
		fuse.FSName("pacman"),
		fuse.Subtype("pacmanfs"),
		fuse.ReadOnly(),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	log.Println("Onwards we go!")

	filesys := &FS{db}
	err = fs.Serve(client, filesys)
	if err != nil {
		log.Fatal(err)
	}

	<-client.Ready

	if err := client.MountError; err != nil {
		log.Fatal(err)
	}

	log.Println("Goodbye.")
}

type FS struct {
	db *alpm.DB
}

var _ fs.FS = (*FS)(nil)

func (filesys FS) Root() (fs.Node, error) {
	root := &InstalledDir{filesys.db}
	return root, nil
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
	contents string
}

var _ = fs.Node(&StupidFile{})
var _ = fs.HandleReadAller(&StupidFile{})

func (file StupidFile) Attr(ctx context.Context, attr *fuse.Attr) error {
	log.Println("StupidFile Attr")

	// XXX figure out some Inode scheme.
	attr.Mode = 0444
	attr.Size = uint64(len(file.contents))
	return nil
}

func (file StupidFile) ReadAll(ctx context.Context) ([]byte, error) {
	log.Println("File ReadAll")
	return []byte(file.contents), nil
}
