package nboci

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/klauspost/compress/zstd"
	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	credentials "github.com/oras-project/oras-credentials-go"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"
)

type PushArgs struct {
	File          []string `arg:"positional,required" help:"boot file"`
	Plain         bool     `arg:"-N,--plain" help:"plain HTTP (insecure)"`
	Repository    string   `arg:"-r,--repository,required" help:"repository (e.g. ghcr.io/user/repo)"`
	Name          string   `arg:"-n,--osname,required" help:"distribution name (e.g. fedora, debian)"`
	Version       string   `arg:"-v,--osversion,required" help:"distribution version (e.g. 45, 9.6)"`
	Architecture  string   `arg:"-a,--osarch,required" help:"architecture (e.g. x86_64, arm64)"`
	Tag           string   `arg:"-t,--tag" help:"tag (default: name-version-arch)"`
	EntryPoint    string   `arg:"-e,--entrypoint,required" help:"entry point (default: shim.efi)"`
	AltEntryPoint string   `arg:"-E,--alt-entrypoint" help:"alternative entry point"`
}

func Push(ctx context.Context, args PushArgs) {
	slog.Debug("checking arguments", "name", args.Name, "version", args.Version, "arch", args.Architecture)
	if !AlphanumRegexp.MatchString(args.Name) {
		Fatal("invalid character in name")
	}
	if !AlphanumRegexp.MatchString(args.Version) {
		Fatal("invalid character in version")
	}
	if !AlphanumRegexp.MatchString(args.Architecture) {
		Fatal("invalid character in architecture")
	}
	if !ArchRegexp.MatchString(args.Architecture) {
		Fatal("unknown architecture")
	}
	if !AlphanumRegexp.MatchString(args.Name) {
		Fatal("invalid character in name")
	}

	// generate tag
	if args.Tag == "" {
		args.Tag = fmt.Sprintf("%s-%s-%s", args.Name, args.Version, args.Architecture)
	}

	repo, err := remote.NewRepository(args.Repository)
	if err != nil {
		FatalErr(err, "cannot create repository")
	}

	repo.Client = &auth.Client{
		Client:     retry.DefaultClient,
		Cache:      auth.DefaultCache,
		Credential: credentials.Credential(NewStore()),
	}
	if args.Plain {
		repo.PlainHTTP = true
	}

	dir := mkTempDir()
	defer os.RemoveAll(dir)

	descs := make([]ocispec.Descriptor, 0, len(args.File))
	for _, f := range args.File {
		Debug("compressing", f)
		a, err := newArtifact(f)
		if err != nil {
			FatalErr(err, "cannot load file")
		}
		d := *a.Descriptor()
		descs = append(descs, d)

		Debug("pushing", f)
		err = repo.Push(ctx, d, a.Reader())
		if err != nil {
			FatalErr(err, "cannot push layer")
		}
	}

	manifest, err := generateManifest(ocispec.DescriptorEmptyJSON,
		args.Name,
		args.Version,
		args.Architecture,
		args.EntryPoint,
		args.AltEntryPoint,
		descs...)
	if err != nil {
		FatalErr(err, "cannot generate manifest")
	}

	desc := content.NewDescriptorFromBytes(ocispec.MediaTypeImageManifest, manifest)

	Print("pushing config")
	err = repo.Push(ctx, ocispec.DescriptorEmptyJSON, bytes.NewReader(ocispec.DescriptorEmptyJSON.Data))
	if err != nil {
		FatalErr(err, "cannot push config")
	}

	Print("pushing manifest")
	err = repo.PushReference(ctx, desc, bytes.NewReader(manifest), args.Tag)
	if err != nil {
		FatalErr(err, "cannot push manifest")
	}
}

type Artifact struct {
	filename  string
	buf       []byte
	srcSize   int64
	srcDigest string
	size      int64
	digest    string
}

func (a *Artifact) Descriptor() *ocispec.Descriptor {
	return &ocispec.Descriptor{
		MediaType: NetbootFileZstdMediaType,
		Digest:    digest.Digest(a.digest),
		Size:      a.size,
		Annotations: map[string]string{
			"org.opencontainers.image.title":     a.filename,
			"org.pulpproject.netboot.src.size":   fmt.Sprintf("%d", a.srcSize),
			"org.pulpproject.netboot.src.digest": a.srcDigest,
		},
	}
}

func (a *Artifact) Reader() io.Reader {
	return bytes.NewReader(a.buf)
}

func compress(in io.Reader, out io.Writer) (string, error) {
	enc, err := zstd.NewWriter(out, zstd.WithEncoderLevel(zstd.SpeedBestCompression))
	if err != nil {
		return "", err
	}
	defer enc.Close()

	hw := sha256.New()
	tr := io.TeeReader(in, hw)

	_, err = io.Copy(enc, tr)
	if err != nil {
		return "", err
	}

	sum := hw.Sum(nil)
	return fmt.Sprintf("sha256:%s", hex.EncodeToString(sum)), nil
}

func newArtifact(file string) (*Artifact, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fs, err := f.Stat()
	if err != nil {
		return nil, err
	}

	srcSize := fs.Size()
	buf := bytes.Buffer{}
	buf.Grow(int(srcSize))
	sd, err := compress(f, &buf)
	if err != nil {
		return nil, err
	}

	buffer := buf.Bytes()
	return &Artifact{
		filename:  filepath.Base(file),
		buf:       buffer,
		size:      int64(len(buffer)),
		digest:    digest.FromBytes(buffer).String(),
		srcSize:   srcSize,
		srcDigest: sd,
	}, nil
}

func generateManifest(config ocispec.Descriptor, name, version, arch, entry, altentry string, layers ...ocispec.Descriptor) ([]byte, error) {
	content := ocispec.Manifest{
		ArtifactType: UnknownArtifactType,
		MediaType:    ocispec.MediaTypeImageManifest,
		Config:       config,
		Layers:       layers,
		Versioned:    specs.Versioned{SchemaVersion: 2},
		Annotations: map[string]string{
			"org.pulpproject.netboot.os.name":       name,
			"org.pulpproject.netboot.os.version":    version,
			"org.pulpproject.netboot.os.arch":       arch,
			"org.pulpproject.netboot.entrypoint":    entry,
			"org.pulpproject.netboot.altentrypoint": altentry,
		},
	}
	return json.Marshal(content)
}
