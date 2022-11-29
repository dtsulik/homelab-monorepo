package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"dagger.io/dagger"
)

var services = []string{"apigw", "intake", "output", "process", "request", "status"}

func main() {
	// TODO move these to env vars

	os.Chdir("../../")
	parent_dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	modified := filter_from_commits(services)

	name_prefix := "dtsulik/gif-doggo-"
	for _, service := range modified {
		fmt.Println("Building and publishing " + service)

		target_dir := "cmd/" + service

		outpath := parent_dir + "/build/ci/"
		err = os.MkdirAll(outpath, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}

		err = build(parent_dir, target_dir, outpath)
		if err != nil {
			log.Fatal(err)
		}

		err = publish(name_prefix+service, outpath)
		if err != nil {
			log.Fatal(err)
		}
		os.RemoveAll(outpath + "/bin/")
		if err != nil {
			log.Fatal(err)
		}
	}
}

func filter_from_commits(files []string) []string {
	// FIXME this looks at last commit, means if several is pushed at once it will miss all but last one
	modified_list, err := exec.Command("git", "log", "--format=", "-n", "1", "--name-only").Output()
	if err != nil {
		log.Fatal(err)
	}

	rv := []string{}
	for _, f := range files {
		if strings.Contains(string(modified_list), f) {
			rv = append(rv, f)
		}
	}
	return rv
}

func build(parent_dir, target_dir, output_dir string) error {
	ctx := context.Background()

	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		return err
	}
	defer client.Close()

	src := client.Host().Directory(parent_dir)

	golang := client.Container().From("golang:1.19.3-alpine")
	golang = golang.WithMountedDirectory("/src", src).
		WithWorkdir("/src").
		WithEnvVariable("CGO_ENABLED", "0")

	path := "output"
	golang = golang.WithExec(
		[]string{"go", "build", "-o", path + "/app", target_dir + "/main.go"},
	)

	_, err = golang.ExitCode(ctx)
	if err != nil {
		return err
	}

	output := golang.Directory(path)
	_, err = output.Export(ctx, output_dir+"/bin/")
	if err != nil {
		return err
	}
	return nil
}

func publish(name, artifact_dir string) error {
	ctx := context.Background()

	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		return err
	}
	defer client.Close()

	src := client.Host().Directory(artifact_dir)
	if err != nil {
		return err
	}

	cn, err := client.Container().Build(src).
		Publish(ctx, name, dagger.ContainerPublishOpts{})

	if err != nil {
		return err
	}

	fmt.Printf("Published: %s", cn)
	return nil
}

// func deploy(name string) error {
// 	return nil
// }
