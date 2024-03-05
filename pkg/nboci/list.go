package nboci

import (
	"context"
	"strings"

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
			if strings.HasSuffix(tag, ".sig") {
				continue
			}

			Print(tag)
		}
		return nil
	})
	if err != nil {
		FatalErr(err, "cannot list tags")
	}
}
