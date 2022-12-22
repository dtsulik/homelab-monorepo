package service

import (
	"ci/internal/artifactfactory"
	"ci/internal/buildfactory"
	"fmt"
	"os"
	"path/filepath"
)

type Service struct {
	Name            string
	Path            string
	ParentPath      string
	artifact        artifactfactory.Artifact
	ReadyForPublish bool
}

type ServiceArtifact struct {
	Name string
	Path string
}

type Artifact interface {
	Publish() error
	Package() error
}

func (s Service) String() string {
	return fmt.Sprintf("Name: %s, Path: %s", s.Name, s.Path)
}

func New(name, ppath, path string) (Service, error) {
	// TODO build type discovery here
	s := Service{
		Name:            name,
		Path:            path,
		ParentPath:      ppath,
		ReadyForPublish: false,
		artifact: artifactfactory.Artifact{
			Name:         name,
			Path:         filepath.Join(os.TempDir(), name),
			BuildFactory: buildfactory.New(buildfactory.Golang),
		},
	}

	return s, nil
}

func (s Service) BuildArtifact() (*artifactfactory.Artifact, error) {
	err := s.artifact.Build(s.ParentPath, s.Path)
	if err != nil {
		return nil, err
	}
	return &s.artifact, nil
}

func (s Service) Artifact() *artifactfactory.Artifact {
	return &s.artifact
}
