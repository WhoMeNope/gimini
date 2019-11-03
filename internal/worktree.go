package internal

import (
	"io"
	"os"
	filepath "path"
	"syscall"
	"time"

	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/osfs"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/filemode"
	"gopkg.in/src-d/go-git.v4/plumbing/format/index"
	"gopkg.in/src-d/go-git.v4/utils/ioutil"
)

type Worktree struct {
	*git.Worktree
	repo             *git.Repository
	systemFilesystem billy.Filesystem
}

func GetWorktree(repo *git.Repository) (Worktree, error) {
	fs := osfs.New("/")

	worktree, err := repo.Worktree()
	if err != nil {
		return Worktree{nil, repo, fs}, err
	}

	return Worktree{worktree, repo, fs}, nil
}

func (w *Worktree) Add(path string) (plumbing.Hash, error) {
	s, err := w.Status()
	if err != nil {
		return plumbing.ZeroHash, err
	}

	idx, err := w.repo.Storer.Index()
	if err != nil {
		return plumbing.ZeroHash, err
	}

	var h plumbing.Hash
	var added bool

	fi, err := w.systemFilesystem.Lstat(path)
	if err != nil || !fi.IsDir() {
		added, h, err = w.doAddFile(idx, s, path)
	} else {
		added, err = w.doAddDirectory(idx, s, path)
	}

	if err != nil {
		return h, err
	}

	if !added {
		return h, nil
	}

	return h, w.repo.Storer.SetIndex(idx)
}

func (w *Worktree) doAddDirectory(idx *index.Index, s git.Status, directory string) (added bool, err error) {
	files, err := w.systemFilesystem.ReadDir(directory)
	if err != nil {
		return false, err
	}

	for _, file := range files {
		name := filepath.Join(directory, file.Name())

		var a bool
		if file.IsDir() {
			a, err = w.doAddDirectory(idx, s, name)
		} else {
			a, _, err = w.doAddFile(idx, s, name)
		}

		if err != nil {
			return
		}

		if !added && a {
			added = true
		}
	}

	return
}

func (w *Worktree) doAddFile(idx *index.Index, s git.Status, path string) (added bool, h plumbing.Hash, err error) {
	if s.File(path).Worktree == git.Unmodified {
		return false, h, nil
	}

	h, err = w.copyFileToStorage(path)
	if err != nil {
		if os.IsNotExist(err) {
			added = true
			h, err = w.deleteFromIndex(idx, path)
		}

		return
	}

	if err := w.addOrUpdateFileToIndex(idx, path, h); err != nil {
		return false, h, err
	}

	return true, h, err
}

func (w *Worktree) copyFileToStorage(path string) (hash plumbing.Hash, err error) {
	fi, err := w.systemFilesystem.Lstat(path)
	if err != nil {
		return plumbing.ZeroHash, err
	}

	obj := w.repo.Storer.NewEncodedObject()
	obj.SetType(plumbing.BlobObject)
	obj.SetSize(fi.Size())

	writer, err := obj.Writer()
	if err != nil {
		return plumbing.ZeroHash, err
	}

	defer ioutil.CheckClose(writer, &err)

	if fi.Mode()&os.ModeSymlink != 0 {
		err = w.fillEncodedObjectFromSymlink(writer, path, fi)
	} else {
		err = w.fillEncodedObjectFromFile(writer, path, fi)
	}

	if err != nil {
		return plumbing.ZeroHash, err
	}

	return w.repo.Storer.SetEncodedObject(obj)
}

func (w *Worktree) fillEncodedObjectFromFile(dst io.Writer, path string, fi os.FileInfo) (err error) {
	src, err := w.systemFilesystem.Open(path)
	if err != nil {
		return err
	}

	defer ioutil.CheckClose(src, &err)

	if _, err := io.Copy(dst, src); err != nil {
		return err
	}

	return err
}

func (w *Worktree) fillEncodedObjectFromSymlink(dst io.Writer, path string, fi os.FileInfo) error {
	target, err := w.systemFilesystem.Readlink(path)
	if err != nil {
		return err
	}

	_, err = dst.Write([]byte(target))
	return err
}

func (w *Worktree) addOrUpdateFileToIndex(idx *index.Index, filename string, h plumbing.Hash) error {
	repoRoot := w.Filesystem.Root()
	repoFilename := filepath.Join(repoRoot, filename)

	e, err := idx.Entry(repoFilename)
	if err != nil && err != index.ErrEntryNotFound {
		return err
	}

	if err == index.ErrEntryNotFound {
		return w.doUpdateFileToIndex(idx.Add(repoFilename), filename, h)
	}
	return w.doUpdateFileToIndex(e, filename, h)
}

func (w *Worktree) doUpdateFileToIndex(e *index.Entry, filename string, h plumbing.Hash) error {
	info, err := w.systemFilesystem.Lstat(filename)
	if err != nil {
		return err
	}

	e.Hash = h
	e.ModifiedAt = info.ModTime()
	e.Mode, err = filemode.NewFromOSFileMode(info.Mode())
	if err != nil {
		return err
	}

	if e.Mode.IsRegular() {
		e.Size = uint32(info.Size())
	}

	fillSystemInfo(e, info.Sys())
	return nil
}

func (w *Worktree) deleteFromIndex(idx *index.Index, path string) (plumbing.Hash, error) {
	repoRoot := w.Filesystem.Root()
	repoPath := filepath.Join(repoRoot, path)

	e, err := idx.Remove(repoPath)
	if err != nil {
		return plumbing.ZeroHash, err
	}

	return e.Hash, nil
}

func fillSystemInfo(e *index.Entry, sys interface{}) {
	if os, ok := sys.(*syscall.Stat_t); ok {
		e.CreatedAt = time.Unix(int64(os.Ctim.Sec), int64(os.Ctim.Nsec))
		e.Dev = uint32(os.Dev)
		e.Inode = uint32(os.Ino)
		e.GID = os.Gid
		e.UID = os.Uid
	}
}
