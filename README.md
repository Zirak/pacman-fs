# pacman-fs

Aims to implement [pkgfs](https://github.com/Zirak/pkgfs) over libalpm (pacman).

## Running

So I decided to be a filthy hipster and use [hy](https://github.com/hylang/hy). Deal with it. Written against hy 0.11 master.

### Building hy
If you want, set up a [virtualenv](http://docs.python-guide.org/en/latest/dev/virtualenvs/). My build was done with python 2.7 (*sigh*).

```sh
git clone https://github.com/hylang/hy.git
cd hy && pip install -e . && cd -
```

### Actually running
```sh
hy main.hy path-to-mountpoint
```

## Layout
`main.hy` is the entry point. It wires things up.

`fshelper.hy` helps it do that.

Nothing else :D

## What's in
- Nothing :D
