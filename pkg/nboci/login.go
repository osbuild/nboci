package nboci

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"

	"oras.land/oras-go/v2/registry/remote/credentials"
	"golang.org/x/term"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
)

type LoginArgs struct {
	Registry string `arg:"positional,required" help:"registry URL"`
	Username string `help:"registry username"`
	Password string `help:"registry password or token"`
}

func Login(ctx context.Context, args LoginArgs) {
	var err error
	reader := bufio.NewReader(os.Stdin)

	if args.Username == "" {
		fmt.Print("Username: ")
		args.Username, err = reader.ReadString('\n')
		if err != nil {
			ErrorErr(err, "cannot read username")
		}
		args.Username = strings.TrimSpace(args.Username)
	}

	if args.Password == "" {
		fmt.Print("Password: ")
		bytePassword, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			ErrorErr(err, "cannot read password")
		}
		args.Password = strings.TrimSpace(string(bytePassword))
		fmt.Printf("\n")
	}

	store := NewStore()
	cred := credential(args.Username, args.Password)
	registry, err := remote.NewRegistry(args.Registry)
	if err != nil {
		FatalErr(err, "cannot create registry")
	}

	if err = credentials.Login(ctx, store, registry, cred); err != nil {
		FatalErr(err, "cannot login")
	}

	Print("Success")
}

func credential(username, password string) auth.Credential {
	if username == "" {
		return auth.Credential{
			RefreshToken: password,
		}
	}
	return auth.Credential{
		Username: username,
		Password: password,
	}
}
