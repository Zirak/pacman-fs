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

	filesys := &FS{handle, db}
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
	handle *alpm.Handle
	db     *alpm.DB
}

var _ fs.FS = (*FS)(nil)

func (filesys FS) Root() (fs.Node, error) {
	root := &InstalledDir{filesys.handle, filesys.db}
	return root, nil
}

type InstalledDir struct {
	handle *alpm.Handle
	db     *alpm.DB
}

var _ = fs.Node(InstalledDir{})

func (InstalledDir) Attr(ctx context.Context, attr *fuse.Attr) error {
	log.Println("Dir Attr")

	attr.Inode = 1
	attr.Mode = os.ModeDir | 0555
	return nil
}

func (dir InstalledDir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	log.Println("Lookup: " + name)

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

	return PkgDir{&pkg}, nil
}

func (dir InstalledDir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	log.Println("ReadDirAll")

	dirs := []fuse.Dirent{}

	for _, pkg := range dir.db.GetPkgcache().Slice() {
		entry := fuse.Dirent{
			Name:  pkg.Name(),
			Type:  fuse.DT_Dir,
		}

		dirs = append(dirs, entry)
	}

	return dirs, nil
}

type PkgDir struct {
	pkg   *alpm.Pkg
}

var pkgDirEntries = []fuse.Dirent{
	{Name: "version", Type: fuse.DT_File},
	{Name: "desc", Type: fuse.DT_File},
}

var _ = fs.Node(PkgDir{})

func (dir PkgDir) Attr(ctx context.Context, attr *fuse.Attr) error {
	log.Println("PkgDir Attr")

	attr.Mode = os.ModeDir | 0555
	return nil
}

func (dir PkgDir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	log.Println("PkgDir Lookup: " + name)


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
		return StupidFile{ dir.pkg.Version() }, nil
	}
	if name == "desc" {
		return StupidFile{ dir.pkg.Desc() }, nil
	}

	return nil, fuse.ENOENT
}

func (dir PkgDir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	return pkgDirEntries, nil
}

// XXX this struct, for some reason, is unhashable, which crashes bazil.
// figure out how to make it hashable and why their examples work.
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
