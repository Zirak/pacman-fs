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

type DBWrapper struct {
	local *alpm.DB
	// an explanation on why we use a slice pointer instead of a slice is
	//available on Amazon for $9.99, limited Christmas edition all year round.
	// It's also available in index-dir.go
	sync *[]*alpm.DB
}

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

	wrapper := DBWrapper{sync: &[]*alpm.DB{}}

	localdb, err := handle.GetLocalDb()
	if err != nil {
		log.Fatal(err)
	}

	wrapper.local = localdb

	for _, syncdbName := range []string{"core", "extra", "community"} {
		syncdb, err := handle.RegisterSyncDb(syncdbName)

		if err != nil {
			log.Fatal(err)
		}

		// wrapper.sync[syncdbName] = syncdb
		sync := append(*wrapper.sync, syncdb)
		wrapper.sync = &sync
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

	filesys := &FS{&wrapper}
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
	dbs *DBWrapper
}

var _ fs.FS = (*FS)(nil)

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
	}, nil
}

func (dir RootDir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	if name == "installed" {
		return InstalledDir{dir.local}, nil
	}

	if name == "index" {
		return IndexDir{dir.sync}, nil
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
