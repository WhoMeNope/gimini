package main

import (
	"fmt"
	"os"

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

	// Print status
	status, err := w.Status()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(status)
}
