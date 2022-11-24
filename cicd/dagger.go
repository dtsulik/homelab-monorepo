package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"dagger.io/dagger"
)

func main() {
	// TODO move these to env vars

	os.Chdir("../services/")
	service_dir, err := os.Getwd()
	os.Chdir("../cicd/")
	if err != nil {
		log.Fatal(err)
	}

	files, err := os.ReadDir(service_dir)
	if err != nil {
		log.Fatal(err)
	}

	name_prefix := "dtsulik/gif-doggo-"
	for _, file := range files {
		fmt.Println("Building and publishing " + file.Name())

		target_dir := service_dir + "/" + file.Name()

		err = build(target_dir, "./")
		if err != nil {
			log.Fatal(err)
		}
		err = publish(name_prefix + file.Name())
		if err != nil {
			log.Fatal(err)
		}
		os.RemoveAll("build/")
		if err != nil {
			log.Fatal(err)
		}
	}
}

func build(target_dir, output_dir string) error {
	ctx := context.Background()

	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		return err
	}
	defer client.Close()

	src := client.Host().Directory(target_dir)

	golang := client.Container().From("golang:1.19.3-alpine")
	golang = golang.WithMountedDirectory("/src", src).
		WithWorkdir("/src").
		WithEnvVariable("CGO_ENABLED", "0")

	path := "build/"
	golang = golang.WithExec(
		[]string{"go", "build", "-o", path + "app"},
	)

	_, err = golang.ExitCode(ctx)
	if err != nil {
		return err
	}

	outpath := filepath.Join(output_dir, path)
	err = os.MkdirAll(outpath, os.ModePerm)
	if err != nil {
		return err
	}

	output := golang.Directory(path)
	_, err = output.Export(ctx, outpath)
	if err != nil {
		return err
	}
	return nil
}

func publish(name string) error {
	ctx := context.Background()

	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		return err
	}
	defer client.Close()

	src := client.Host().Directory(".")
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
