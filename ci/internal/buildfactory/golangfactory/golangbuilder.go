package golangfactory

import (
	"context"
	"path/filepath"

	"dagger.io/dagger"
)

type GolangFactory struct {
}

func (g GolangFactory) Test() error {
	return nil
}

func (g GolangFactory) Build(parentPath, targetPath, outputPath string) error {

	ctx := context.Background()

	client, err := dagger.Connect(ctx) // dagger.WithLogOutput(os.Stdout)
	if err != nil {
		return err
	}
	defer client.Close()

	src := client.Host().Directory(parentPath)

	golang := client.Container().From("golang:1.19.3-alpine")
	golang = golang.WithMountedDirectory("/src", src).
		WithWorkdir(filepath.Join("/src", targetPath)).
		WithEnvVariable("CGO_ENABLED", "0")

	path := "/tmp/output/"
	golang = golang.WithExec(
		[]string{"go", "build", "-o", path, "./..."},
	)

	_, err = golang.ExitCode(ctx)
	if err != nil {
		return err
	}

	output := golang.Directory(path)
	_, err = output.Export(ctx, outputPath)
	if err != nil {
		return err
	}
	return nil
}
