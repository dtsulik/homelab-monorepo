package git

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

func getChangeName(change *object.Change) string {
	var empty = object.ChangeEntry{}
	if change.From != empty {
		return change.From.Name
	}

	return change.To.Name
}

func getCommit(r *git.Repository, lastn int) ([]*object.Commit, error) {
	until := time.Now()
	cIter, err := r.Log(&git.LogOptions{Until: &until})
	if err != nil {
		return nil, err
	}

	var commits []*object.Commit
	for i := 0; i < lastn; i++ {
		c, err := cIter.Next()
		if err != nil {
			return nil, err
		}
		commits = append(commits, c)
	}
	return commits, nil
}

func GetChangedFiles() ([]string, error) {
	r, err := GetRepo()
	if err != nil {
		return nil, err
	}
	commits, err := getCommit(r, 2)
	if err != nil {
		return nil, err
	}
	return getChanges(commits[0], commits[1])
}

// FIXME this breaks with more than 2 commits per push
func getChanges(curr, prev *object.Commit) ([]string, error) {
	ctree, err := curr.Tree()
	if err != nil {
		return nil, err
	}
	ptree, err := prev.Tree()
	if err != nil {
		return nil, err
	}

	// TODO check if patch is more useful
	// patch, err := ctree.Patch(ptree)
	// if err != nil {
	// 	return nil, err
	// }
	// fmt.Println("----- Patch Stats ------")
	// for _, fileStat := range patch.Stats() {
	// 	fmt.Println("file:", fileStat.Name)
	// 	changedFiles = append(changedFiles, fileStat.Name)
	// }
	// fmt.Println(changedFiles)

	changes, err := ctree.Diff(ptree)
	if err != nil {
		return nil, err
	}
	// fmt.Println(changes)
	var changedFiles []string
	for _, change := range changes {
		// Ignore deleted files
		// action, err := change.Action()
		// if err != nil {
		// 	return err
		// }
		// if action == merkletrie.Delete {
		// 	//fmt.Println("Skipping delete")
		// 	continue
		// }
		name := getChangeName(change)
		changedFiles = append(changedFiles, name)
	}

	return changedFiles, err
}

func GetRepo() (*git.Repository, error) {
	directory := "."

	r, err := git.PlainOpenWithOptions(directory, &git.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		return nil, err
	}
	return r, nil
}

func DoRelease(r *git.Repository) error {
	token, ok := os.LookupEnv("GH_TOKEN")
	if ok {
		return errors.New("GH_TOKEN not set")
	}

	w, err := r.Worktree()
	if err != nil {
		return err
	}
	directory := "."

	log.Println("echo \"hello world!\" > example-git-file")
	filename := filepath.Join(directory, "example-git-file")
	err = os.WriteFile(filename, []byte("hello world!"), 0644)
	if err != nil {
		return err
	}

	log.Println("git add example-git-file")
	_, err = w.Add("example-git-file")
	if err != nil {
		return err
	}

	log.Println("git status --porcelain")
	status, err := w.Status()
	if err != nil {
		return err
	}

	fmt.Println(status)

	log.Println("git commit -m \"example go-git commit\"")
	commit, err := w.Commit("example go-git commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "David Tsulaia",
			Email: "dtsulik@gmail.com",
			When:  time.Now(),
		},
	})

	if err != nil {
		return err
	}

	log.Println("git show -s")
	obj, err := r.CommitObject(commit)
	if err != nil {
		return err
	}

	fmt.Println(obj)

	log.Println("git push")
	err = r.Push(&git.PushOptions{
		Auth: &http.BasicAuth{
			Username: "dtsulik@gmail.com",
			Password: token,
		},
	})
	return err
}
