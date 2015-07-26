# pacman-fs

Aims to implement [pkgfs](https://docs.google.com/document/d/1Fi1ebe_rAq4v-JNW8i2IbT4iUHIPro-wbVT86tBhW14/edit#heading=h.y92gnqagqz2j) over libalpm (pacman)

## Running

```sh
go run *.go path-to-mountpoint
```

I forgot whether go gets dependencies too and all that, so in case it doesn't:

```sh
go get bazil.org/fuse
go get bazil.org/fuse/fs
```

### pacman-fs in action
[![pacman-fs in action, via asciinema](https://asciinema.org/a/e8bik65jepshaagi9thyaxjgf.png)](https://asciinema.org/a/e8bik65jepshaagi9thyaxjgf)

## Layout
`main.go` is the entry point. It's just in charge of parsing arguments and wiring our filesystem to FUSE.

Each of the subdirectories has its own README for your pleasure, but a concise description:
* In the `src` directory you'll find the filesystem itself, with all its directories and galore.
* libalpm (Arch Linux Package Manager library) is what pacman serves as a frontend of. The `alpm` directory is my wrapping of (some) of its features.

So far, there isn't so much to the code. I expect this will change as the project ages.

## What's in
- Mountpoint acts as `/pkg`, featuring both `index/` and `installed/`!
- Each package has a description, version and size files.
- Packages have a `deps` folder with symlinks to their dependencies.
