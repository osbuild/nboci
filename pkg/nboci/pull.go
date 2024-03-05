package nboci

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/klauspost/compress/zstd"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	credentials "github.com/oras-project/oras-credentials-go"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"
)

type PullArgs struct {
	Source      string `arg:"positional,required" help:"repository:tag" placeholder:"REPOSITORY:{TAG|DIGEST}"`
	Destination string `arg:"-d,--destination" default:"." help:"destination directory (default: pwd)" placeholder:"DIRECTORY"`
}

func Pull(ctx context.Context, args PullArgs) {
	// check if destination is valid
	if _, err := os.Stat(args.Destination); os.IsNotExist(err) {
		err = os.MkdirAll(args.Destination, 0700)
		if err != nil {
			Fatal("cannot create destination directory")
		}
	}

	fs, err := file.New(args.Destination)
	if err != nil {
		panic(err)
	}
	defer fs.Close()

	var onlyTag string
	ss := strings.SplitN(args.Source, ":", 2)
	if len(ss) >= 2 {
		onlyTag = ss[1]
	}

	repo, err := remote.NewRepository(ss[0])
	if err != nil {
		panic(err)
	}

	repo.Client = &auth.Client{
		Client:     retry.DefaultClient,
		Cache:      auth.NewCache(),
		Credential: credentials.Credential(NewStore()),
	}

	err = repo.Tags(ctx, "", func(tags []string) error {
		for _, tag := range tags {
			if (onlyTag != "" && tag != onlyTag) || (strings.HasSuffix(tag, ".sig")) {
				continue
			}

			desc, err := repo.Resolve(ctx, tag)
			if err != nil {
				return err
			}

			if desc.MediaType == ocispec.MediaTypeImageManifest {
				Debug("processing", tag)

				blob, err := content.FetchAll(ctx, repo, desc)
				if err != nil {
					return err
				}
				var manifest ocispec.Manifest
				if err := json.Unmarshal(blob, &manifest); err != nil {
					return err
				}

				destPath, err := makePath(manifest.Annotations)
				if err != nil {
					continue
				}
				dirname := path.Join(args.Destination, destPath)

				ss, err := content.Successors(ctx, repo, desc)
				if err != nil {
					FatalErr(err, "cannot list successors")
				}

				for _, s := range ss {
					if s.MediaType != NetbootFileZstdMediaType {
						continue
					}

					name, ok := s.Annotations["org.opencontainers.image.title"]
					if !ok {
						Fatal("artifact is missing org.opencontainers.image.title annotation for", s.Digest.String())
					}

					err := os.MkdirAll(dirname, 0777)
					if err != nil {
						FatalErr(err, "cannot create destination directory")
					}
					filename := path.Join(dirname, name)

					fdigest, _ := fileDigest(filename)
					rdigest, ok := s.Annotations["org.pulpproject.netboot.src.digest"]
					if ok && rdigest == fdigest {
						Debug("digest match for", filename)
						continue
					}

					// download
					Print("downloading", filename)
					hash, err := download(ctx, repo, s, filename)
					if err != nil {
						FatalErr(err, "cannot download", filename)
					}

					if rdigest != "" && rdigest != hash {
						Fatal("downloaded file", filename, "has different digest", hash, "than expected", rdigest)
					}
				}

				ep := manifest.Annotations["org.pulpproject.netboot.entrypoint"]
				ensureEntrypoint(path.Join(dirname, "boot"), path.Join(dirname, filepath.Base(ep)))
				aep := manifest.Annotations["org.pulpproject.netboot.altentrypoint"]
				ensureEntrypoint(path.Join(dirname, "boot-alt"), path.Join(dirname, filepath.Base(aep)))
			}
		}
		return nil
	})
	if err != nil {
		FatalErr(err, "cannot list tags")
	}
}

func ensureEntrypoint(link, dest string) {
	if dest == "" {
		return
	}

	if _, err := os.Stat(dest); errors.Is(err, os.ErrNotExist) {
		ErrorErr(err, "entrypoint destination not exist")
	}

	orig, err := os.Readlink(link)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		// does not exist
		makeSymlink(link, dest)
		return
	} else if err != nil {
		// not a symlink
		ErrorErr(err, "cannot create entrypoint symlink")
	}

	if filepath.Base(orig) != filepath.Base(dest) {
		// is different symlink
		err = os.Remove(link)
		if err != nil {
			FatalErr(err, "cannot remove existing file", link)
		}
		makeSymlink(link, dest)
	}
}

func makeSymlink(link, dest string) {
	Debug("updating entrypoint", filepath.Base(link), "->", filepath.Base(dest))
	err := os.Symlink(filepath.Base(dest), link)
	if err != nil {
		ErrorErr(err, "cannot create symlink")
	}
}

func makePath(a map[string]string) (string, error) {
	keys := []string{
		"org.pulpproject.netboot.os.name",
		"org.pulpproject.netboot.os.version",
		"org.pulpproject.netboot.os.arch",
	}

	for _, key := range keys {
		if _, ok := a[key]; !ok {
			return "", fmt.Errorf("missing %s annotation", key)
		}
	}

	return path.Join(a[keys[0]], a[keys[1]], a[keys[2]]), nil
}

func fileDigest(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	sum := hash.Sum(nil)
	return fmt.Sprintf("sha256:%s", hex.EncodeToString(sum)), nil
}

func download(ctx context.Context, repo *remote.Repository, desc ocispec.Descriptor, dest string) (string, error) {
	rc, err := repo.Fetch(ctx, desc)
	if err != nil {
		return "", err
	}

	r, err := zstd.NewReader(rc)
	if err != nil {
		return "", err
	}
	defer r.Close()

	w, err := os.Create(dest)
	if err != nil {
		return "", err
	}
	defer w.Close()

	hw := sha256.New()
	tr := io.TeeReader(r, hw)

	_, err = io.Copy(w, tr)
	if err != nil {
		return "", err
	}

	sum := hw.Sum(nil)
	return fmt.Sprintf("sha256:%s", hex.EncodeToString(sum)), nil
}
