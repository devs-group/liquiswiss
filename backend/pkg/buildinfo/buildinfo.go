// Package buildinfo carries the version identifiers stamped into the binary at
// build time. CI passes them via -ldflags -X, so a running container can always
// answer "what exactly is deployed" through GET /api/config.
package buildinfo

var (
	// GitSHA is the commit the image was built from.
	GitSHA = "dev"
	// Version is the semver tag the image was released under (or "dev").
	Version = "dev"
	// AppSHA fingerprints only the app-relevant sources, so a deploy that
	// touched nothing the app runs can be recognised as such.
	AppSHA = ""
)
