package version

import "fmt"

// BuildVersion build verion
var BuildVersion = "0.1."

// BuildNum build number
var BuildNum string

// BuildBy build user
var BuildBy string

// BuildAt build time
var BuildAt string

// GoVersion go version
var GoVersion string

// Commit commit
var Commit string

// New create version object
func New() Version {
	if BuildNum == "" {
		BuildNum = "1"
	}
	return Version{Version: fmt.Sprintf("%s%s", BuildVersion, BuildNum), BuildAt: BuildAt, BuildNum: BuildNum, BuildBy: BuildBy, GoVersion: GoVersion, Commit: Commit}
}

// Version verion object
type Version struct {
	Version   string
	BuildNum  string
	BuildBy   string
	BuildAt   string
	GoVersion string
	Commit    string
}
