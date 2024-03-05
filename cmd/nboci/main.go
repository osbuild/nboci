package main

import (
	"context"

	"github.com/lzap/oci-netboot/pkg/nboci"

	arg "github.com/alexflint/go-arg"
)

type args struct {
	Login   *nboci.LoginArgs  `arg:"subcommand:login" help:"login to registry"`
	Logout  *nboci.LogoutArgs `arg:"subcommand:logout" help:"logout from registry"`
	Push    *nboci.PushArgs   `arg:"subcommand:push" help:"push files to registry"`
	List    *nboci.PullArgs   `arg:"subcommand:list" help:"list available tags in registry"`
	Pull    *nboci.PullArgs   `arg:"subcommand:pull" help:"pull files to registry"`
	Verbose bool
}

func (a args) Version() string {
	return "nboci 1.0.0"
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
	} else if args.Push != nil {
		nboci.Push(ctx, *args.Push)
	} else if args.Pull != nil {
		nboci.Pull(ctx, *args.Pull)
	} else {
		parser.Fail("unknown subcommand")
	}
}
