package helm

import (
	"io"
	"os"

	"dagger/pkg/semver"

	"gopkg.in/yaml.v3"
)

func TestChart() {

	f, err := os.OpenFile("Chart2.yaml", os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	c := HelmChart{}

	err = c.Read(f)
	if err != nil {
		panic(err)
	}

	dep := c.Dependency("app-template")
	if dep != nil {
		dep.Version = semver.Semver(dep.Version).BumpVersion(semver.Patch).String()
	}
	c.Version = semver.Semver(c.Version).BumpVersion(semver.Patch).String()
	c.Write(os.Stdout)

	err = f.Truncate(0)
	if err != nil {
		panic(err)
	}
	f.Seek(0, 0)
	err = c.Write(f)
	if err != nil {
		panic(err)
	}
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

func (c *HelmChart) Dependency(name string) *HelmChartDependency {
	deps := c.dependencies()
	for i, dep := range *deps {
		if dep.Name == name {
			return &(*deps)[i]
		}
	}
	return nil
}

func (c *HelmChart) Read(r io.Reader) error {

	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal([]byte(data), &c)
	if err != nil {
		return err
	}
	return nil
}

func (c *HelmChart) String() string {
	d, err := yaml.Marshal(c)
	if err != nil {
		return "error"
	}
	return string(d)
}

func (c *HelmChart) dependencies() *[]HelmChartDependency {
	return &c.Dependencies
}

func (c *HelmChart) Write(w io.Writer) error {
	d, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	_, err = w.Write(d)
	if err != nil {
		return err
	}

	return nil
}
