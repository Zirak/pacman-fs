import fuse
import errno, stat

# TODO should probably be removed
import logging
logging.basicConfig(level=logging.INFO, format='[%(created)d] %(message)s')

class Operations(fuse.Operations):
    '''Small helper class which potentially defers read calls to read_all,
    from which you need only return a string/buffer once, and it will do the
    offset/size calculations for you! hurray!
    '''
    def read(self, rest, size, offset, *args):
        # wtf
        if rest:
            raise FuseOSError(errno.ENOTDIR)

        if hasattr(self, 'read_all'):
            if not hasattr(self, '_read_cache'):
                self._read_cache = self.read_all(*args)

            return self._read_cache[offset:offset+size]

        raise FuseOSError(errno.EIO)

class Directory(Operations, fuse.LoggingMixIn):
    def __init__(self, **children):
        self.handlers = {}

        for childname, handler in children.items():
            self.handle(childname, handler)

    # "router" stuff

    def handle(self, sub_path, handler):
        self.handlers[sub_path] = handler
        return self

    def match(self, path_parts):
        print('match:', path_parts)

        # when we've reached the bottom of the recursion
        if not path_parts:
            return (self, ())

        sub_path = path_parts[0]
        if sub_path in self.handlers:
            handler = self.handlers[sub_path]
            rest = path_parts[1:]

            # sub-directory
            if hasattr(handler, 'match'):
                return handler.match(rest)

            return (handler, rest)

        return (None, ())

    # FUSE stuff
    def __call__(self, op, path, *args):
        print('--------------------------')
        logging.info('%s on %s: %s', op, path, args)

        if path == '/':
            route, rest = (self, ())
        else:
            route, rest = self.match(tuple(path.strip('/').split('/')))

        if route is None:
            print('Not found')
            raise fuse.FuseOSError(errno.ENOENT)

        args = (rest,) + args
        if hasattr(route, op):
            return getattr(route, op)(*args)
        # regular Operations / weirdo with __call__
        elif hasattr(route, '__call__'):
            return route(op, *args)

    def getattr(self, rest, *args):
        print('getattr', rest, args)

        return dict(st_mode=(stat.S_IFDIR | 0o755))

    def readdir(self, *args):
        print('readdir', args)
        return self.handlers.keys()

class StupidFile(Operations):
    def __init__(self, contents):
        self.contents = contents

    def getattr(self, *args):
        return dict(st_mode=(stat.S_IFREG | 0o644),
                    st_size=len(self.contents))

    def read_all(self, *args):
        return self.contents
