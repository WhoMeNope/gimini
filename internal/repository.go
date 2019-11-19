package internal

import (
  "os"

	"gopkg.in/src-d/go-billy.v4/osfs"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/storage/filesystem"
	"gopkg.in/src-d/go-git.v4/plumbing/cache"
)

const repoPath string = "/.gimini"

type Repository struct {
  git.Repository

  config config
}

func OpenOrInit() (*Repository, error) {
  // construct repo path
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	repoPathFull := home + repoPath

	// init repo if does not exist
	os.MkdirAll(repoPathFull, os.ModeDir|0777)
	fs := osfs.New(repoPathFull)
	st := filesystem.NewStorage(fs, cache.NewObjectLRUDefault())
	plainRepo, err := git.Init(st, fs)
	if err != nil && err != git.ErrRepositoryAlreadyExists {
		return nil, err
	}
	// open the repo
	plainRepo, err = git.PlainOpen(repoPathFull)
	if err != nil {
		return nil, err
	}

  // get config
  config, err := getConfig()
	if err != nil {
		return nil, err
	}
  config.save()

  return &Repository{*plainRepo, config}, nil
}

