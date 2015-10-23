(import fuse fshelper)
(import alpm)

(defclass InstalledDir [fshelper.Directory]
  (defn --init-- [self db]
    (setv self.db db))

  (defn match [self path-parts &rest args]
    (if (not path-parts)
      (, self path-parts)

      (let [pkg (-> self.db (.get-local) (.get-package (car path-parts)))]
        (if pkg
          (.match (InstalledPkgDir pkg) (cdr path-parts))

          (, nil (,))))))

  (defn readdir [self path &rest args]
    (map (lambda [x] x.name)
         (-> self.db (.get-local) (.get-packages)))))

(defclass InstalledPkgDir [fshelper.Directory]
  (defn --init-- [self pkg]
    (setv self.pkg pkg)
    (setv self.children {})
    (.gen-children self))

  (defn gen-children [self]
    (.update self.children
             { "version" (fshelper.StupidFile self.pkg.version)
               "description" (fshelper.StupidFile self.pkg.description)
               "size" (fshelper.StupidFile (str self.pkg.install-size)) })))

(defmain [dollarzero mntdir]
  (setv root (fshelper.Directory))
  (setv db-wrapper (.create-db-wrapper alpm))

  (root.handle "installed" (InstalledDir db-wrapper))

  (root.handle "sync" (fshelper.StupidFile "#!/bin/sh
pacman -Sy
" :mode 0o555))

  (root.handle "upgrade" (fshelper.StupidFile "#!/bin/sh
pacman -Su
" :mode 0o555))

  (fuse.FUSE root mntdir :foreground True :allow-other True))
