package project

import (
	"ci/internal/service"
	"ci/pkg/git"
	"fmt"
)

// FIXME all of the channel methods have a bug where
// if one service fails rest of go routines are not killed
// use context to cancel all go routines in case of error
func (p *Project) Cleanup() error {
	errc := make(chan error)
	statusc := make(chan string)

	fmt.Println("::group::Cleaning up")
	for _, s := range p.services {
		go func(svc service.Service) {
			a := svc.Artifact()

			err := a.Cleanup()
			if err != nil {
				errc <- err
				return
			}
			statusc <- fmt.Sprintf("Cleaned up %s", svc.Name)
		}(s)
	}

	for i := 0; i < len(p.services); i++ {
		select {
		case err := <-errc:
			return err
		case s := <-statusc:
			fmt.Println(s)
		}
	}
	fmt.Println("::endgroup::")
	return nil
}

func (p *Project) Build() error {
	fmt.Println("::group::Detecting changes")
	files, err := git.GetChangedFiles()
	if err != nil {
		return err
	}
	fmt.Println("Changed files: ", files)
	fmt.Println("::endgroup::")

	// TODO implement filterChanges
	filteredServices, filteredManifests := p.filterChanges(files)

	errc := make(chan error)
	statusc := make(chan string)

	fmt.Println("::group::Building services")
	for _, s := range filteredServices {
		go func(svc service.Service) {
			_, err := svc.BuildArtifact()
			if err != nil {
				errc <- err
				return
			}
			statusc <- fmt.Sprintf(`{"service": "%s", "state": "built", "error": ""}`, svc.Name)
		}(s)
	}

	for i := 0; i < len(filteredServices); i++ {
		select {
		case err := <-errc:
			return err
		case s := <-statusc:
			fmt.Println(s)
		}
	}
	fmt.Println("::endgroup::")

	fmt.Println("::group::Building manifests")
	for _, m := range filteredManifests {
		fmt.Println("changed manifest: ", m.path)
	}
	fmt.Println("::endgroup::")

	return nil
}

func (p *Project) Publish() error {
	fmt.Println("::group::Publishing artifacts")

	errc := make(chan error)
	statusc := make(chan string)

	filteredServices := []service.Service{}
	for _, s := range p.services {
		if !s.ReadyForPublish {
			continue
		}
		filteredServices = append(filteredServices, s)
		go func(svc service.Service) {
			a := svc.Artifact()
			err := a.Publish("docker.io/dtsulik/" + p.Name + "-" + svc.Name + ":latest")
			if err != nil {
				errc <- err
				return
			}
			statusc <- fmt.Sprintf(`{"service": "%s", "digest": "%s", "state": "published", "error": ""}`, svc.Name, a.Digest)
		}(s)
	}

	for i := 0; i < len(filteredServices); i++ {
		select {
		case err := <-errc:
			return err
		case s := <-statusc:
			fmt.Println(s)
		}
	}
	fmt.Println("::endgroup::")
	return nil
}
