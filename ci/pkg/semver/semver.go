package semver

import (
	"fmt"
	"regexp"
	"strconv"
)

var semver_regex string = `v?([0-9]+)\.([0-9]+)?\.([0-9]+)?`

type VersionPosition int

const (
	Major VersionPosition = iota
	Minor
	Patch
)

type Semver string

func (v Semver) String() string {
	return string(v)
}

// TODO add support
func (v Semver) BumpVersion(vpos VersionPosition) Semver {
	re := regexp.MustCompile(semver_regex)

	vs := string(v)

	if !re.Match([]byte(vs)) {
		return v
	}
	res := re.FindStringSubmatch(vs)
	if len(res) != 4 {
		return v
	}
	vmajor, err := strconv.Atoi(res[1])
	if err != nil {
		return v
	}
	vminor, err := strconv.Atoi(res[2])
	if err != nil {
		return v
	}

	vpatch, err := strconv.Atoi(res[3])
	if err != nil {
		return v
	}
	switch vpos {
	case Major:
		vmajor += 1
		vminor = 0
		vpatch = 0
	case Minor:
		vminor += 1
		vpatch = 0
	case Patch:
		vpatch += 1
	}
	vs = fmt.Sprintf("%d.%d.%d", vmajor, vminor, vpatch)
	return Semver(vs)
}
