package main

import (
	"./alpm"
	"./src"

	"flag"
	"fmt"
	"log"
	"os"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

func main() {
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Pass a mountpoint, dummy.")
		os.Exit(1)
	}

	mountpoint := flag.Arg(0)

	handle, err := alpm.Init()
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Release()

	localdb, err := handle.GetLocalDb()
	if err != nil {
		log.Fatal(err)
	}

	syncdbs := []*alpm.DB{}
	for _, syncdbName := range []string{"core", "extra", "community"} {
		syncdb, err := handle.RegisterSyncDb(syncdbName)

		if err != nil {
			log.Fatal(err)
		}

		syncdbs = append(syncdbs, syncdb)
	}

	client, err := fuse.Mount(
		mountpoint,
		fuse.FSName("pacman"),
		fuse.Subtype("pacmanfs"),
		fuse.AllowOther(),
		fuse.ReadOnly(),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	log.Println("Onwards we go!")

	filesys := pacmanfs.NewFilesystem(localdb, &syncdbs)
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
