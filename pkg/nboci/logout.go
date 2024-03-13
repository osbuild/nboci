package nboci

import (
	"context"

	"oras.land/oras-go/v2/registry/remote/credentials"
)

type LogoutArgs struct {
	Registry string `arg:"positional,required" help:"registry URL"`
}

func Logout(ctx context.Context, args LogoutArgs) {
	store := NewStore()
	if err := credentials.Logout(ctx, store, args.Registry); err != nil {
		FatalErr(err, "cannot logout")
	}

	Print("Success")
}
