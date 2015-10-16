(import [fuse [FUSE]] [fshelper [Directory StupidFile]])

(defmain [dollarzero mntdir]
  (setv root (Directory))
  (root.handle "hai" (StupidFile "hai"))
  (root.handle "somedir" (Directory :file (StupidFile "file!\n")))

  (FUSE root mntdir :foreground True))
