(import ctypes ctypes.util)
(import [collections [namedtuple]])

(setv alpm (get ctypes.cdll (ctypes.util.find-library "alpm")))

;; annoying declarations

(setv handle-t ctypes.c-voidp)
(setv db-t ctypes.c-voidp)
(setv pkg-t ctypes.c-voidp)
(setv error-t ctypes.c-int)

(defclass PointerList [ctypes.Structure]
  (defn --iter-- [self]
    (yield self.data)
    (setv nextp self.next)
    (when (and nextp nextp.contents)
      (yield-from nextp.contents))))

(setv PointerList.-fields-
      [(, "data" ctypes.c-voidp)
       (, "prev" (ctypes.POINTER PointerList))
       (, "next" (ctypes.POINTER PointerList))])

;; ... kill me

;; alpm_handle_t *alpm_initialize(const char *root, const char *dbpath, alpm_errno_t *err);
(setv alpm.alpm-initialize.restype handle-t)

;; alpm_errno_t alpm_errno(alpm_handle_t *handle);
(setv alpm.alpm-errno.argtypes [handle-t])

;; alpm_db_t *alpm_get_localdb(alpm_handle_t *handle);
(setv alpm.alpm-get-localdb.argtypes [handle-t])
(setv alpm.alpm-get-localdb.restype db-t)

;; alpm_list_t *alpm_db_get_pkgcache(alpm_db_t *db);
(setv alpm.alpm-db-get-pkgcache.argtypes [db-t])
(setv alpm.alpm-db-get-pkgcache.restype (ctypes.POINTER PointerList))

;; alpm_pkg_t *alpm_db_get_pkg(alpm_db_t *db, const char *name);
(setv alpm.alpm-db-get-pkg.argtypes [db-t ctypes.c-char-p])
(setv alpm.alpm-db-get-pkg.restype pkg-t)

;; const char *alpm_pkg_get_name(alpm_pkg_t *pkg);
(setv alpm.alpm-pkg-get-name.argtypes [pkg-t])
(setv alpm.alpm-pkg-get-name.restype ctypes.c-char-p)

;; const char *alpm_pkg_get_version(alpm_pkg_t *pkg);
(setv alpm.alpm-pkg-get-version.argtypes [pkg-t])
(setv alpm.alpm-pkg-get-version.restype ctypes.c-char-p)

;; const char *alpm_pkg_get_desc(alpm_pkg_t *pkg);
(setv alpm.alpm-pkg-get-desc.argtypes [pkg-t])
(setv alpm.alpm-pkg-get-desc.restype ctypes.c-char-p)

;; off_t alpm_pkg_get_isize(alpm_pkg_t *pkg);
(setv alpm.alpm-pkg-get-isize.argtypes [pkg-t])
(setv alpm.alpm-pkg-get-isize.restype ctypes.c-uint)

;; const char *alpm_strerror(alpm_errno_t err);
(setv alpm.alpm-strerror.restype ctypes.c-char-p)

(defclass Package [(namedtuple "Package" ["ptr" "name" "version"
                                          "description" "install_size"])]
  (with-decorator staticmethod
    (defn from-ptr [pkgptr]
      (Package :ptr pkgptr
               :name (alpm.alpm-pkg-get-name pkgptr)
               :version (alpm.alpm-pkg-get-version pkgptr)
               :description (alpm.alpm-pkg-get-desc pkgptr)
               :install-size (alpm.alpm-pkg-get-isize pkgptr)))))

(defclass DBWrapper [object]
  (defn --init-- [self handle]
    (setv self.handle handle))

  (defn get-local [self]
    ;; TODO caching logic :/
    (unless (hasattr self "localdb")
      (setv self.localdb (DB (alpm.alpm-get-localdb self.handle))))
    self.localdb))

(defclass DB [object]
  (defn --init-- [self dbptr]
    (setv self.ptr dbptr))

  (defn get-packages [self]
    (let [pkglist-ptr (alpm.alpm-db-get-pkgcache self.ptr)]
      (map Package.from-ptr pkglist-ptr.contents)))

  (defn get-package [self name]
    (setv pkgptr (alpm.alpm-db-get-pkg self.ptr name))
    (when pkgptr
      (Package.from-ptr pkgptr))))

(defn create-db-wrapper []
  (let [err (ctypes.c_int)
        hand (alpm.alpm-initialize (str "/") (str "/var/lib/pacman") (ctypes.byref err))]
    (DBWrapper hand)))
