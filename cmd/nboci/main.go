package main

import (
	"context"
	"log/slog"

	"github.com/lzap/oci-netboot/pkg/nboci"

	arg "github.com/alexflint/go-arg"
)

type args struct {
	Login   *nboci.LoginArgs `arg:"subcommand:login" help:"login to registry"`
	Push    *nboci.PushArgs  `arg:"subcommand:push" help:"push files to registry"`
	Pull    *nboci.PullArgs  `arg:"subcommand:pull" help:"pull files to registry"`
	Quiet   bool
	Verbose bool
	Debug   bool
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

	if args.Debug {
		nboci.InitLogger(slog.LevelDebug)
	} else if args.Verbose {
		nboci.InitLogger(slog.LevelInfo)
	} else if args.Quiet {
		nboci.InitLogger(slog.LevelError)
	} else {
		nboci.InitLogger(slog.LevelWarn)
	}

	if args.Login != nil {
		nboci.Login(ctx, *args.Login)
	} else if args.Push != nil {
		nboci.Push(ctx, *args.Push)
	} else if args.Pull != nil {
		nboci.Pull(ctx, *args.Pull)
	} else {
		parser.Fail("unknown subcommand")
	}
}
