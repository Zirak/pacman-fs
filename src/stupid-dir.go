package pacmanfs

import (
	"../alpm"

	"strings"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context"
)

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
