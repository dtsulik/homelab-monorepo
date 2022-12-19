package service

import "fmt"

type Service struct {
	Name string
	Path string
}

func (s Service) String() string {
	return fmt.Sprintf("Name: %s, Path: %s", s.Name, s.Path)
}

func New(name, path string) (Service, error) {
	return Service{
		Name: name,
		Path: path,
	}, nil
}

func (s Service) Build() error {
	return nil
}
