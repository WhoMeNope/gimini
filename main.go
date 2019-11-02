package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/src-d/go-billy.v4/osfs"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/cache"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/storage/filesystem"
)

const repoDir string = "/.gimini"

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Init if does not exist
	os.MkdirAll(home+repoDir, os.ModeDir|0777)
	fs := osfs.New(home + repoDir)
	st := filesystem.NewStorage(fs, cache.NewObjectLRUDefault())
	repo, err := git.Init(st, nil)
	if err != nil && err != git.ErrRepositoryAlreadyExists {
		fmt.Println(err)
		return
	}
	// Open the repo
	repo, err = git.PlainOpen(home + repoDir)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Walk throufh the dir
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
		}

		return nil
	}
	err = filepath.Walk(os.Args[1], ff)
	if err != nil {
		fmt.Printf("error walking the path %q: %v\n", os.Args[1], err)
	}

	// Length of the HEAD history
	fmt.Println("git rev-list HEAD --count")
	// ... retrieving the HEAD reference
	ref, err := repo.Head()
	if err != nil {
		fmt.Println(err)
		return
	}
	// ... retrieves the commit history
	cIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		fmt.Println(err)
		return
	}
	// ... just iterates over the commits
	var cCount int
	err = cIter.ForEach(func(c *object.Commit) error {
		cCount++
		return nil
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(cCount)

}
