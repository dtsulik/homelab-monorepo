package helm

import (
	"fmt"
	"io"
	"os"

	"dagger/pkg/semver"

	"gopkg.in/yaml.v3"
)

func TestChart() {
	c := HelmChart{}
	err := c.Read("Chart.yaml")
	if err != nil {
		fmt.Println(err)
	}
	deps := c.dependencies()
	for i, dep := range *deps {
		dep.Version = semver.Semver(dep.Version).BumpVersion(semver.Patch).String()
		(*deps)[i] = dep
	}
	c.Version = semver.Semver(c.Version).BumpVersion(semver.Patch).String()
	c.dump(os.Stdout)
	c.Write("Chart.yaml")
}

type HelmChartDependency struct {
	Name       string `yaml:"name"`
	Version    string `yaml:"version"`
	Repository string `yaml:"repository"`
	Condition  string `yaml:"condition"`
}

type HelmChart struct {
	ApiVersion   string                `yaml:"apiVersion"`
	AppVersion   string                `yaml:"appVersion"`
	Description  string                `yaml:"description"`
	Name         string                `yaml:"name"`
	Version      string                `yaml:"version"`
	Dependencies []HelmChartDependency `yaml:"dependencies"`
}

func (c *HelmChart) Read(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal([]byte(data), &c)
	if err != nil {
		return err
	}
	return nil
}

func (c *HelmChart) dump(d *os.File) {
	io.WriteString(d, fmt.Sprintln(c))
}

func (c *HelmChart) dependencies() *[]HelmChartDependency {
	return &c.Dependencies
}

func (c *HelmChart) Write(file string) error {
	d, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(file, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(d)
	if err != nil {
		return err
	}

	return nil
}
