package nboci

import "regexp"

const UnknownArtifactType = "application/vnd.unknown.artifact.v1"
const EmptyType = "application/vnd.oci.empty.v1+json"
const MediaType = "application/x-netboot-file+zstd"

var AlphanumRegexp regexp.Regexp
var ArchRegexp regexp.Regexp

func init() {
	AlphanumRegexp = *regexp.MustCompile(`^[a-z0-9\._]*$`)
	ArchRegexp = *regexp.MustCompile(`^(x86_64|aarch64|ppc64|ppc64le)$`)
}
