package buildfactory

import (
	"ci/internal/buildfactory/golangfactory"
	"ci/internal/buildfactory/makefilefactory"
)

type BuildFactory interface {
	Test() error
	Build(parentPath, inputPath, outputPath string) error
}

type BuildType int

const (
	Golang BuildType = 1 << iota
	Makefile
	// Java
)

// TODO maybe not default and use errors
// or return some abstract unimplemented factory
func New(t BuildType) BuildFactory {
	switch t {
	case Golang:
		return golangfactory.GolangFactory{}
	case Makefile:
		return makefilefactory.MakefileFactory{}
	default:
		return makefilefactory.MakefileFactory{}
	}
}
