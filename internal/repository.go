package internal

import (
  "os"

	"gopkg.in/src-d/go-billy.v4/osfs"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/storage/filesystem"
	"gopkg.in/src-d/go-git.v4/plumbing/cache"
)

const repo string = "/.gimini"

func OpenOrInit() (*git.Repository, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	repoPath := home + repo

	// Init if does not exist
	os.MkdirAll(repoPath, os.ModeDir|0777)
	fs := osfs.New(repoPath)
	st := filesystem.NewStorage(fs, cache.NewObjectLRUDefault())
	repo, err := git.Init(st, fs)
	if err != nil && err != git.ErrRepositoryAlreadyExists {
		return nil, err
	}
	// Open the repo
	repo, err = git.PlainOpen(repoPath)
	if err != nil {
		return nil, err
	}

  return repo, nil
}

