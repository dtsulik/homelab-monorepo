package main

import (
	"fmt"
	"log"
	"os"

	"ci/internal/project"
)

func main() {
	os.Chdir("..")
	proj, err := project.New("gif-doggo", ".")
	if err != nil {
		log.Fatal("Project setup failed: ", err)
	}
	fmt.Println("Project setup complete:\n", proj)
	err = proj.Build()
	if err != nil {
		log.Fatal("Project build failed: ", err)
	}
}
