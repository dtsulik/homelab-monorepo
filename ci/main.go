package main

import (
	"log"
	"os"

	"ci/internal/project"
)

func main() {
	os.Chdir("..")
	proj, err := project.New("gif-doggo", "projects/gif-doggo")
	if err != nil {
		log.Fatal("Project setup failed: ", err)
	}
	// fmt.Println("Project setup complete:\n", proj)
	err = proj.Build()
	if err != nil {
		log.Fatal("Project build failed: ", err)
	}
	err = proj.Publish()
	if err != nil {
		log.Fatal("Project publish failed: ", err)
	}
	err = proj.Cleanup()
	if err != nil {
		log.Fatal("Project cleanup failed: ", err)
	}
}
