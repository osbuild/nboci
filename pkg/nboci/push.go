package nboci

import (
	"fmt"
	"log/slog"
	"os"
	"path"
)

type PushArgs struct {
	File         []string `arg:"positional,required" help:"boot file"`
	Repository   string   `arg:"-r,--repository,required" help:"repository (e.g. ghcr.io/user/repo)"`
	Name         string   `arg:"-n,--name,required" help:"distribution name (e.g. fedora, debian)"`
	Version      string   `arg:"-v,--version,required" help:"distribution version (e.g. 45, 9.6)"`
	Architecture string   `arg:"-a,--arch,required" help:"architecture (e.g. x86_64, arm64)"`
	Tag          string   `arg:"-t,--tag" help:"tag (default: name-version-arch)"`
}

func Push(args PushArgs) {
	// check arguments
	slog.Debug("checking arguments", "name", args.Name, "version", args.Version, "arch", args.Architecture)
	if !AlphanumRegexp.MatchString(args.Name) {
		ExitWithErrorMsg("invalid character in name")
	}
	if !AlphanumRegexp.MatchString(args.Version) {
		ExitWithErrorMsg("invalid character in version")
	}
	if !AlphanumRegexp.MatchString(args.Architecture) {
		ExitWithErrorMsg("invalid character in architecture")
	}
	if !ArchRegexp.MatchString(args.Architecture) {
		ExitWithErrorMsg("unknown architecture")
	}
	if !AlphanumRegexp.MatchString(args.Name) {
		ExitWithErrorMsg("invalid character in name")
	}

	// generate tag
	if args.Tag == "" {
		args.Tag = fmt.Sprintf("%s-%s-%s", args.Name, args.Version, args.Architecture)
	}

	// create a temporary dir
	dir, err := os.MkdirTemp("", "oci-netboot-")
	if err != nil {
		ExitWithError("Cannot create temp directory", err)
	}
	defer os.RemoveAll(dir)

	// copy and compress files into the temporary dir
	ofiles := make([]string, 0, len(args.File))
	for _, f := range args.File {
		dest := path.Join(dir, path.Base(f))
		slog.Debug("compressing file", "from", f, "to", dest)
		err := Command("zstd", "-9", "-q", f, "-o", dest)
		if err != nil {
			ExitWithError("zstd compressor returned error", err)
		}

		ofiles = append(ofiles, fmt.Sprintf("%s:%s", path.Base(f), MediaType))
	}

	// switch to the temp directory
	slog.Debug("switching to temp directory", "dir", dir)
	pwd, err := os.Getwd()
	if err != nil {
		ExitWithError("cannot get workding directory", err)
	}
	os.Chdir(dir)
	defer os.Chdir(pwd)

	// call oras
	oras := []string{
		"push",
		fmt.Sprintf("%s:%s", args.Repository, args.Tag),
		"--config", fmt.Sprintf("/dev/null:%s", MediaType),
		"--artifact-type", ArtifactType,
	}
	oras = append(oras, ofiles...)
	err = ORAS(oras...)
	if err != nil {
		ExitWithError("oras push returned an error", err)
	}
}