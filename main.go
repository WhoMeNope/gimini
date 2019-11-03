package main

import (
	"fmt"
	"os"

	"github.com/WhoMeNope/gimini/internal"

	"gopkg.in/src-d/go-billy.v4/osfs"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/cache"
	"gopkg.in/src-d/go-git.v4/storage/filesystem"
)

const repo string = "/.gimini"

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(err)
		return
	}
	repoPath := home + repo

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
	w, err := internal.GetWorktree(repo)
	if err != nil {
		fmt.Println(err)
		return
	}

	hash, err := w.Add(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(hash)

	status, err := w.Status()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(status)
}
