package artifactfactory

import (
	"ci/internal/buildfactory"
	"context"
	"os"

	"dagger.io/dagger"
)

type ArtifactFactory interface {
	Publish() error
	Cleanup() error
	Build(parentPath, inputPath string) (string, error)
	Test() error
}

type Artifact struct {
	Name         string
	Path         string
	Digest       string
	BuildFactory buildfactory.BuildFactory
}

func (a *Artifact) Publish(fullname string) error {
	ctx := context.Background()

	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		return err
	}
	defer client.Close()

	src := client.Host().Directory(a.Path)

	cn, err := client.Container().From("alpine").
		WithMountedDirectory("/tmp/app/", src).
		// WithExec([]string{"addgroup", "-S", "1000", "&&", "adduser", "-HS", "1000", "-G", "1000"}).
		WithExec([]string{"mkdir", "/app/"}).
		WithExec([]string{"cp", "-r", "/tmp/app/" + a.Name, "/app/"}).
		WithExec([]string{"chown", "-R", "nobody:nobody", "/app/"}).
		WithUser("nobody").
		WithEntrypoint([]string{"/app/" + a.Name}).
		Publish(ctx, fullname, dagger.ContainerPublishOpts{})

	a.Digest = cn

	if err != nil {
		return err
	}
	return nil
}

func (a *Artifact) Package() error {
	return nil
}

func (a *Artifact) Build(parentPath, inputPath string) error {
	a.BuildFactory.Build(parentPath, inputPath, a.Path)
	return nil
}

func (a *Artifact) Cleanup() error {
	return os.RemoveAll(a.Path)
}

func (a *Artifact) Test() error {
	return nil
}
