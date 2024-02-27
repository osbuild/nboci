package nboci

import "context"

type LoginArgs struct {
	Username string `help:"registry username"`
	Password string `help:"registry password or token"`
}

func Login(ctx context.Context, args LoginArgs) {
	var err error
	if args.Username == "" {
		err = ORAS("login", "--password", args.Password)
	} else {
		err = ORAS("login", "--username", args.Username, "--password", args.Password)
	}

	if err != nil {
		ExitWithError("oras failed to login", err)
	}
}
