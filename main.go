package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/src-d/go-billy.v4/osfs"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/cache"
	"gopkg.in/src-d/go-git.v4/storage/filesystem"
)

const repoDir string = "/.gimini/.git"

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(err)
		return
	}

	repoPath := home + repoDir

	// Init if does not exist
	os.MkdirAll(repoPath, os.ModeDir|0777)
	fs := osfs.New(repoPath)
	st := filesystem.NewStorage(fs, cache.NewObjectLRUDefault())
	repo, err := git.Init(st, fs)
	if err != nil && err != git.ErrRepositoryAlreadyExists {
		fmt.Println(err)
		return
	}
	// Open the repo
	repo, err = git.PlainOpen(repoPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	w, err := repo.Worktree()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Walk through the dir
	var ff = func(pathX string, infoX os.FileInfo, errX error) error {

		// first thing to do, check error. and decide what to do about it
		if errX != nil {
			fmt.Printf("error 「%v」 at a path 「%q」\n", errX, pathX)
			return errX
		}

		// find out if it's a dir or file, if file, print info
		if !infoX.IsDir() {
			fmt.Printf("%v\n", pathX)
			// fmt.Printf("  dir: 「%v」\n", filepath.Dir(pathX))
			// fmt.Printf("  file name 「%v」\n", infoX.Name())
			// fmt.Printf("  extenion: 「%v」\n", filepath.Ext(pathX))

			_, err = w.Add(pathX)
			if err != nil {
				fmt.Println(err)
				return err
			}

		}

		return nil
	}
	err = filepath.Walk(os.Args[1], ff)
	if err != nil {
		fmt.Printf("error walking the path %q: %v\n", os.Args[1], err)
		return
	}

	status, err := w.Status()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(status)
}
