package internal

import (
	"bytes"
	"strings"
  "fmt"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/format/index"
	"gopkg.in/src-d/go-git.v4/utils/merkletrie"
	"gopkg.in/src-d/go-git.v4/utils/merkletrie/noder"
  mindex "gopkg.in/src-d/go-git.v4/utils/merkletrie/index"
)

func (w *Worktree) Status() (git.Status, error) {
	var hash plumbing.Hash

	ref, err := w.repo.Head()
	if err != nil && err != plumbing.ErrReferenceNotFound {
		return nil, err
	}

	if err == nil {
		hash = ref.Hash()
	}

	return w.status(hash)
}

func (w *Worktree) status(commit plumbing.Hash) (git.Status, error) {
	s := make(git.Status)

	// left, err := w.diffCommitWithStaging(commit, false)
	// if err != nil {
	// 	return nil, err
	// }

	// for _, ch := range left {
	// 	a, err := ch.Action()
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	fs := s.File(nameFromAction(&ch))
	// 	fs.Worktree = git.Unmodified

	// 	switch a {
	// 	case merkletrie.Delete:
	// 		s.File(ch.From.String()).Staging = git.Deleted
	// 	case merkletrie.Insert:
	// 		s.File(ch.To.String()).Staging = git.Added
	// 	case merkletrie.Modify:
	// 		s.File(ch.To.String()).Staging = git.Modified
	// 	}
	// }

	right, err := w.diffStagingWithWorktree()
	if err != nil {
		return nil, err
	}

	for _, ch := range right {
		a, err := ch.Action()
		if err != nil {
			return nil, err
		}

		fs := s.File(nameFromAction(&ch))
		if fs.Staging == git.Untracked {
			fs.Staging = git.Unmodified
		}

		switch a {
		case merkletrie.Delete:
			fs.Worktree = git.Deleted
		case merkletrie.Insert:
			fs.Worktree = git.Untracked
			fs.Staging = git.Untracked
		case merkletrie.Modify:
			fs.Worktree = git.Modified
		}
	}

	return s, nil
}

func nameFromAction(ch *merkletrie.Change) string {
	name := ch.To.String()
	if name == "" {
		return ch.From.String()
	}

	return name
}

func (w *Worktree) diffStagingWithWorktree() (merkletrie.Changes, error) {
	idx, err := w.repo.Storer.Index()
	if err != nil {
		return nil, err
	}

	// Translate repo paths to system paths
	repoRoot := w.Filesystem.Root()
	for _, idx_entry := range idx.Entries {
		idx_entry.Name = strings.TrimPrefix(idx_entry.Name, repoRoot)
	}

  fmt.Println(idx)

	// Compare with system files
  paths := w.repo.config.Paths
	pathNodeMap := w.repo.config.getFilesystemNodes(w.systemFilesystem)

  pathIndexMap := make(map[string]*index.Index)
  for _, prefix := range paths {
    pathIndexMap[prefix] = &index.Index{}

    for _, idx_entry := range idx.Entries {
      if strings.HasPrefix(idx_entry.Name, prefix) {
        pathIndexMap[prefix].Entries = append(pathIndexMap[prefix].Entries, idx_entry)
      }
    }
  }

  var changes merkletrie.Changes
  for _, prefix := range paths {
    fmt.Println(prefix)

    from := mindex.NewRootNode(pathIndexMap[prefix])
    to := pathNodeMap[prefix]

    fmt.Println(from.Name)
    fmt.Println(to.Name)

    c, err := merkletrie.DiffTree(from, to, diffTreeIsEquals)
    if err != nil {
      return nil, err
    }

    changes = append(changes, c...)
  }

	return changes, nil
}

func (w *Worktree) diffCommitWithStaging(commit plumbing.Hash, reverse bool) (merkletrie.Changes, error) {
	var t *object.Tree
	if !commit.IsZero() {
		c, err := w.repo.CommitObject(commit)
		if err != nil {
			return nil, err
		}

		t, err = c.Tree()
		if err != nil {
			return nil, err
		}
	}

	return w.diffTreeWithStaging(t, reverse)
}

func (w *Worktree) diffTreeWithStaging(t *object.Tree, reverse bool) (merkletrie.Changes, error) {
	var from noder.Noder
	if t != nil {
		from = object.NewTreeRootNode(t)
	}

	idx, err := w.repo.Storer.Index()
	if err != nil {
		return nil, err
	}

	to := mindex.NewRootNode(idx)

	if reverse {
		return merkletrie.DiffTree(to, from, diffTreeIsEquals)
	}

	return merkletrie.DiffTree(from, to, diffTreeIsEquals)
}

var emptyNoderHash = make([]byte, 24)

// diffTreeIsEquals is a implementation of noder.Equals, used to compare
// noder.Noder, it compare the content and the length of the hashes.
//
// Since some of the noder.Noder implementations doesn't compute a hash for
// some directories, if any of the hashes is a 24-byte slice of zero values
// the comparison is not done and the hashes are take as different.
func diffTreeIsEquals(a, b noder.Hasher) bool {
	hashA := a.Hash()
	hashB := b.Hash()

	if bytes.Equal(hashA, emptyNoderHash) || bytes.Equal(hashB, emptyNoderHash) {
		return false
	}

	return bytes.Equal(hashA, hashB)
}
