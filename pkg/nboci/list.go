package nboci

import (
	"context"
	"encoding/json"
	"fmt"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"
)

type ListArgs struct {
	Source string `arg:"positional,required" help:"repository" placeholder:"REPOSITORY"`
}

func List(ctx context.Context, args ListArgs) {
	repo, err := remote.NewRepository(args.Source)
	if err != nil {
		panic(err)
	}

	repo.Client = &auth.Client{
		Client: retry.DefaultClient,
		Cache:  auth.NewCache(),
	}

	err = repo.Tags(ctx, "", func(tags []string) error {
		for _, tag := range tags {
			desc, err := repo.Resolve(ctx, tag)
			if err != nil {
				return err
			}

			if desc.MediaType == ocispec.MediaTypeImageManifest {
				blob, err := content.FetchAll(ctx, repo, desc)
				if err != nil {
					return err
				}
				var manifest ocispec.Manifest
				if err := json.Unmarshal(blob, &manifest); err != nil {
					return err
				}

				Debug("Annotations", fmt.Sprintf("%v", manifest.Annotations))

				if _, err := makePath(manifest.Annotations); err == nil {
					Print(tag)
				}
			}
		}
		return nil
	})
	if err != nil {
		FatalErr(err, "cannot list tags")
	}
}
