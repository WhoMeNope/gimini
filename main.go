package main

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"

	"github.com/WhoMeNope/gimini/internal"
)

func main() {
	// Open repo (init if does not exist)
	repo, err := internal.OpenOrInit()
	if err != nil {
		fmt.Println(err)
		return
	}

	w, err := internal.GetWorktree(repo)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Add dir
	hash, err := w.Add(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(hash)

	// Commit
	hash, err = w.Commit("Commit message", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "gimini",
			Email: "gimini@acme.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(hash)

	// Get commit
	commit, err := object.GetCommit(w.Repo().Storer, hash)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(commit)

	// Get tree
	tree, err := object.GetTree(w.Repo().Storer, commit.TreeHash)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(tree)

	// Print status
	status, err := w.Status()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(status)
}
