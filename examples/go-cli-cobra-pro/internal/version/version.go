package version

import "fmt"

var (
	// Version is set at build time.
	Version = "dev"
	// Commit is the git commit sha set at build time.
	Commit = ""
	// Date is the build date set at build time.
	Date = ""
)

// Human returns a human-readable version string.
func Human() string {
	if Commit != "" && Date != "" {
		return fmt.Sprintf("%s (%s, %s)", Version, Commit, Date)
	}
	if Commit != "" {
		return fmt.Sprintf("%s (%s)", Version, Commit)
	}
	return Version
}
