(import fuse)
(import errno stat)

(import logging)
(logging.basicConfig :level logging.INFO :format "[%(created)d] %(message)s")

(defclass Operations [fuse.Operations]
  "Small helper class which potentially defers read calls to read_all,
  from which you need only return a string/buffer once, and it will do the
  offset/size calculations for you! hurray!
  "

  (defn read [self rest size offset &rest args]
    (when rest
      (raise (FuseOSError errno.ENOTDIR)))

    (unless (hasattr self "read_all")
      (raise (FuseOSError errno.EIO)))

    (unless (hasattr self "_read_cache")
      (setv self.-read-cache (apply self.read-all args)))

    (cut self.-read-cache offset (+ offset size))))

(defclass Directory [Operations fuse.LoggingMixIn]
  (defn --init-- [self &kwargs children]
    (setv self.children (or children {})))

  ;; "router" stuff
  (defn handle [self sub-path handler]
    (assoc self.children sub-path handler))

  (defn match [self path-parts]
    (print "match:" path-parts)

    (if path-parts
      (if (in (car path-parts) self.children)
        (do
         ;; this was supposed to be a let block, but some versions of hy act
         ;;weird with multi-var `let`
         (setv handler (get self.children (car path-parts)))
         (setv rest (cdr path-parts))

         (if (hasattr handler "match")
           ;; a sub-directory, hurray
           (.match handler rest)

           ;; or maybe just a regular child
           (, handler rest)))
        ;; dunno what to do with this one
        (, nil (,)))

      ;; reached bottom of recursion, sacrifice ourselves
      (, self (,))))

  (defn --call-- [self op path &rest args]
    (print "--------------------------")
    (logging.info "%s on %s: %s" op path args)

    (setv (, route rest) (if (= path "/")
                           (, self (,))
                           (self.match (tuple (-> path (.strip "/") (.split "/"))))))
    (when (is route None)
      (print "404 %s not found" path)
      (raise (fuse.FuseOSError errno.ENOENT)))

    (setv args (+ (, rest) args))

    (cond
     [(hasattr route op)
      (apply (getattr route op) args)]
     [(hasattr route "__call__")
      (apply route args)]))

  (defn getattr [self rest &rest args]
    (print "getattr" rest args)
    ;; the hy emacs mode was having trouble with |, so you'll have to excuse
    ;; the magic. it's: (| stat.S_IFDIR 0o755)
    (setv mode 16877)
    (dict :st-mode mode))

  (defn readdir [self &rest args]
    (print "readdir" args self.children)
    (.keys self.children)))

(defclass StupidFile [Operations]
  (defn --init-- [self contents]
    (setv self.contents contents))

  (defn getattr [self &rest args]
    ;; again, excuse the French
    ;; (| stat.S_IFREG 0o644)
    (setv mode 33412)
    (dict :st-mode mode :st-size (len self.contents)))

  (defn read-all [self &rest args]
    self.contents))
