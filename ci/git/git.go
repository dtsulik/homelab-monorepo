package git

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/utils/merkletrie"

	"gopkg.in/yaml.v3"
)

var semver_regex string = `v?([0-9]+)\.([0-9]+)?\.([0-9]+)?`

func readYaml(file string) (map[interface{}]interface{}, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	m := make(map[interface{}]interface{})
	err = yaml.Unmarshal([]byte(data), &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func writeYaml(file string, data map[interface{}]interface{}) error {
	d, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(file, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(d)
	if err != nil {
		log.Fatalln("asdf", err)
		return err
	}
	return nil
}

func bumpSemverPatch(v string) (string, error) {
	re := regexp.MustCompile(semver_regex)

	if !re.Match([]byte(v)) {
		return "", errors.New("invalid version")
	}
	res := re.FindStringSubmatch(v)
	if len(res) != 4 {
		return "", errors.New("invalid version")
	}
	patch, err := strconv.Atoi(res[3])
	if err != nil {
		return "", err
	}
	v = fmt.Sprintf("%s.%s.%d", res[1], res[2], patch+1)

	return v, nil
}

var filename string = "./Chart.yaml"

func ReadCharts() {
	y, err := readYaml(filename)
	if err != nil {
		log.Fatal(err)
	}

	if version, ok := y["version"]; ok {
		v := fmt.Sprintf("%v", version)
		updated_v, err := bumpSemverPatch(v)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Updated version: %s\n", updated_v)
		y["version"] = updated_v

		if err := writeYaml(filename, y); err != nil {
			log.Fatal(err)
		}
	}
}

func getChangeName(change *object.Change) string {
	var empty = object.ChangeEntry{}
	if change.From != empty {
		return change.From.Name
	}

	return change.To.Name
}

func GetCommit(r *git.Repository) error {
	since := time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC)
	until := time.Date(2023, 7, 30, 0, 0, 0, 0, time.UTC)
	cIter, err := r.Log(&git.LogOptions{Since: &since, Until: &until})
	if err != nil {
		return err
	}

	c, err := cIter.Next()
	if err != nil {
		return err
	}
	cprev, err := cIter.Next()
	if err != nil {
		return err
	}
	fmt.Println(c)
	fmt.Println(cprev)

	ctree, err := c.Tree()
	if err != nil {
		return err
	}
	ptree, err := cprev.Tree()
	if err != nil {
		return err
	}
	patch, err := ctree.Patch(ptree)
	if err != nil {
		return err
	}
	fmt.Println("----- Patch Stats ------")
	var changedFiles []string
	for _, fileStat := range patch.Stats() {
		fmt.Println(fileStat.Name)
		changedFiles = append(changedFiles, fileStat.Name)
	}
	fmt.Println(changedFiles)

	changes, err := ctree.Diff(ptree)
	if err != nil {
		return err
	}
	fmt.Println(changes)
	fmt.Println("----- Changes -----")
	for _, change := range changes {
		// Ignore deleted files
		action, err := change.Action()
		if err != nil {
			return err
		}
		if action == merkletrie.Delete {
			//fmt.Println("Skipping delete")
			continue
		}
		// Get list of involved files
		name := getChangeName(change)
		fmt.Println(name)
	}

	return nil
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
