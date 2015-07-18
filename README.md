# pacman-fs

Aims to implement [pkgfs](https://docs.google.com/document/d/1Fi1ebe_rAq4v-JNW8i2IbT4iUHIPro-wbVT86tBhW14/edit#heading=h.y92gnqagqz2j) over libalpm (pacman)

## Running

```sh
go run main.go path-to-mountpoint
```

I forget what if go gets dependencies too and all that, so in case it doesn't:

```sh
go get bazil.org/fuse
go get bazil.org/fuse/fs
```

## Layout
libalpm (Arch Linux Package Manager library) is what pacman serves as a frontend of. The `alpm` directory is my wrapping of (some) of its features.

Currently there's only one `main.go` (which you need to run) which builds the filesystem and all. Don't worry, it'll be split up and happiness all over.

## What's in
- Mountpoint currently acts as `/pkg/installed`
- Each package has a description and version files
