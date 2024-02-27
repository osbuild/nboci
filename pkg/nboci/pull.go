package nboci

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	oras "oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"
)

type PullArgs struct {
	Source      string `arg:"positional,required" help:"repository:tag" placeholder:"REPOSITORY:{TAG|DIGEST}"`
	Destination string `arg:"-d,--destination" default:"." help:"destination directory (default: .)" placeholder:"DIRECTORY"`
}

func Pull(ctx context.Context, args PullArgs) {
	// check if destination is valid
	if _, err := os.Stat(args.Destination); os.IsNotExist(err) {
		err = os.MkdirAll(args.Destination, 0700)
		if err != nil {
			ExitWithErrorMsg("cannot create destination directory")
		}
	}

	fs, err := file.New(args.Destination)
	if err != nil {
		panic(err)
	}
	defer fs.Close()

	var tags []string
	ss := strings.SplitN(args.Source, ":", 2)
	if len(ss) < 2 {
		// todo - detect all tags from a repo
		tags = append(tags, "latest")
	} else {
		tags = []string{ss[1]}
	}

	repo, err := remote.NewRepository(ss[0])
	if err != nil {
		panic(err)
	}

	repo.Client = &auth.Client{
		Client: retry.DefaultClient,
		Cache:  auth.NewCache(),
	}

	for _, tag := range tags {
		slog.DebugContext(ctx, "pulling tag", "tag", tag)
		manifestDescriptor, err := oras.Copy(ctx, repo, tag, fs, tag, oras.DefaultCopyOptions)
		if err != nil {
			panic(err)
		}
		fmt.Println("manifest descriptor:", manifestDescriptor)
	}

}
