package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"ci/internal/project"

	"dagger.io/dagger"
)

// List of files modification of which should trigger rebuild
var services = []string{"apigw", "intake", "output", "process", "request", "random", "status"}

func main() {
	current_dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	parent_dir := filepath.Dir(current_dir)
	// fmt.Println("Parent dir: " + parent_dir)

	os.Chdir("..")
	proj, err := project.New("gif-doggo", ".")
	if err != nil {
		log.Fatal("Project setup failed: ", err)
	}
	// fmt.Println("Project setup complete:\n", proj)
	proj.Build()
	os.Exit(0)

	name_prefix := "dtsulik/gif-doggo-"
	for _, service := range services {
		log.Println("Building and publishing " + service)

		target_dir := "cmd/" + service

		outpath := parent_dir + "/ci/"
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

	cn, err := client.Container().
		Build(src).
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
