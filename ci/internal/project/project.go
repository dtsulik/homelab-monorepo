package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ci/pkg/helm"

	"ci/internal/service"
)

type Manifest struct {
	path  string
	chart *helm.HelmChart
}

type Project struct {
	Name             string
	Path             string
	services         []service.Service
	projectManifest  Manifest
	serviceManifests []Manifest
	libraryManifests []Manifest
}

func (m Manifest) String() string {
	return fmt.Sprintf("Path: %s\n%s", m.path, m.chart)
}

func (p *Project) String() string {
	s := fmt.Sprintf("Project: %s, Path: %s\n", p.Name, p.Path)
	for _, service := range p.services {
		s += fmt.Sprintf("Service: %s\n", service)
	}
	s += fmt.Sprintf("Project Manifest:\n%s\n", p.projectManifest)
	for _, manifest := range p.serviceManifests {
		s += fmt.Sprintf("Service Manifests:\n%s\n", manifest)
	}
	for _, manifest := range p.libraryManifests {
		s += fmt.Sprintf("Library Manifests:\n%s\n", manifest)
	}

	return s
}

func New(name, path string) (*Project, error) {
	dirs, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	// checks
	manifestExists := false
	cmdExists := false

	p := &Project{
		Name: name,
		Path: path,
	}

	for _, dir := range dirs {
		if dir.Name() == "manifests" {
			err := p.setupManifests(filepath.Join(path, dir.Name()))
			if err != nil {
				return nil, err
			}
			manifestExists = true
		}
		if dir.Name() == "cmd" {
			err := p.setupServices(filepath.Join(path, dir.Name()))
			if err != nil {
				return nil, err
			}
			cmdExists = true
		}
	}

	if !manifestExists {
		return nil, fmt.Errorf("manifests directory not found")
	}
	if !cmdExists {
		return nil, fmt.Errorf("cmd directory not found")
	}

	return p, nil
}

func (p *Project) setupManifests(path string) error {
	manifestDirs, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	// checks
	projectManifestExists := false
	serviceManifestExists := false

	for _, manifestDir := range manifestDirs {
		if manifestDir.Name() == p.Name {
			projectManifestExists = true
			p.projectManifest.path = filepath.Join(path, manifestDir.Name(), "Chart.yaml")
			p.projectManifest.chart, err = helm.NewFromFile(p.projectManifest.path)
			if err != nil {
				return err
			}
		}
		if manifestDir.Name() == "services" {
			serviceManifestExists = true
			serviceManifestDirs, err := os.ReadDir(filepath.Join(path, manifestDir.Name()))
			if err != nil {
				return err
			}
			for _, serviceManifestDir := range serviceManifestDirs {
				if serviceManifestDir.IsDir() {
					m := Manifest{
						path: filepath.Join(path, manifestDir.Name(), serviceManifestDir.Name(), "Chart.yaml"),
					}
					m.chart, err = helm.NewFromFile(m.path)
					if err != nil {
						return err
					}
					p.serviceManifests = append(p.serviceManifests, m)
				}
			}
		}
		if manifestDir.Name() == "library" {
			libraryManifestDirs, err := os.ReadDir(filepath.Join(path, manifestDir.Name()))
			if err != nil {
				return err
			}
			for _, libraryManifestDir := range libraryManifestDirs {
				if libraryManifestDir.IsDir() {
					m := Manifest{
						path: filepath.Join(path, manifestDir.Name(), libraryManifestDir.Name(), "Chart.yaml"),
					}
					m.chart, err = helm.NewFromFile(filepath.Join(path, manifestDir.Name(),
						libraryManifestDir.Name(), "Chart.yaml"))

					if err != nil {
						return err
					}
					p.libraryManifests = append(p.libraryManifests, m)
				}
			}
		}
	}

	if !projectManifestExists {
		return fmt.Errorf("project manifest not found")
	}
	if !serviceManifestExists {
		return fmt.Errorf("service manifest not found")
	}

	return nil
}

func (p *Project) setupServices(path string) error {
	serviceDirs, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	for _, serviceDir := range serviceDirs {
		if serviceDir.IsDir() {
			s, err := service.New(serviceDir.Name(), p.Path, filepath.Join(path, serviceDir.Name()))
			if err != nil {
				return err
			}
			p.services = append(p.services, s)
		}
	}
	return nil
}

// TODO aaaand generics go here
func (p *Project) filterChanges(files []string) ([]service.Service, []Manifest) {
	var services []service.Service
	var manifests []Manifest

	for _, file := range files {
		for _, s := range p.services {
			if strings.HasPrefix(file, s.Path) {
				services = append(services, s)
			}
		}
		for _, m := range p.serviceManifests {
			if strings.HasPrefix(file, m.path) {
				manifests = append(manifests, m)
			}
		}
		for _, m := range p.libraryManifests {
			if strings.HasPrefix(file, m.path) {
				manifests = append(manifests, m)
			}
		}
	}

	return services, manifests
}
