package version

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// VersionCmd version command
var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version",
	Run: func(cmd *cobra.Command, args []string) {
		v := New()
		content, err := json.Marshal(v)
		if err != nil {
			fmt.Printf("Cannot get version information: %v\n", err)
			return
		}
		fmt.Printf("%s\n", string(content))
	},
}

// BuildVersion build verion
var BuildVersion = "EXPERIMENTAL"

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
	return Version{Version: BuildVersion, BuildAt: BuildAt, BuildBy: BuildBy, GoVersion: GoVersion, Commit: Commit}
}

// Version verion object
type Version struct {
	Version   string
	BuildBy   string
	BuildAt   string
	GoVersion string
	Commit    string
}
