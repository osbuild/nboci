package main

import (
	"context"
	"runtime/debug"
	"strings"

	arg "github.com/alexflint/go-arg"
	"github.com/lzap/nboci/pkg/nboci"
)

type args struct {
	Login   *nboci.LoginArgs  `arg:"subcommand:login" help:"login to registry"`
	Logout  *nboci.LogoutArgs `arg:"subcommand:logout" help:"logout from registry"`
	Push    *nboci.PushArgs   `arg:"subcommand:push" help:"push files to registry"`
	List    *nboci.ListArgs   `arg:"subcommand:list" help:"list available tags in registry"`
	Pull    *nboci.PullArgs   `arg:"subcommand:pull" help:"pull files to registry"`
	Verbose bool
}

func (a args) Version() string {
	str := strings.Builder{}
	str.WriteString("nboci " + nboci.BuildCommit + "\n" + nboci.BuildGoVersion + " " + nboci.BuildTime + "\n")

	if bi, ok := debug.ReadBuildInfo(); ok {
		for _, dep := range bi.Deps {
			if strings.Contains(dep.Path, "oras") || strings.Contains(dep.Path, "sigstore") {
				str.WriteString(dep.Path + " " + dep.Version + "\n")
			}
		}
	}

	return str.String()
}

func main() {
	ctx := context.Background()
	var args args
	parser := arg.MustParse(&args)
	if parser.Subcommand() == nil {
		parser.Fail("missing subcommand")
	}

	if args.Verbose {
		nboci.Verbose = true
	}

	if args.Login != nil {
		nboci.Login(ctx, *args.Login)
	} else if args.Logout != nil {
		nboci.Logout(ctx, *args.Logout)
	} else if args.List != nil {
		nboci.List(ctx, *args.List)
	} else if args.Push != nil {
		nboci.Push(ctx, *args.Push)
	} else if args.Pull != nil {
		nboci.Pull(ctx, *args.Pull)
	} else {
		parser.Fail("unknown subcommand")
	}
}
