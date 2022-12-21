package makefilefactory

import "fmt"

type MakefileFactory struct {
}

func (m MakefileFactory) Test() error {
	fmt.Println("Testing with makefile")
	return nil
}

func (m MakefileFactory) Build(parentPath, inputPath, outputPath string) error {
	fmt.Println("Building with makefile")
	return nil
}
