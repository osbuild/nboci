package nboci

import (
	"log/slog"
	"strings"

	"oras.land/oras/cmd/oras/root"
)

func ORAS(args ...string) error {
	cmd := root.New()
	cmd.SetArgs(args)
	slog.Debug("executing oras", "cmd", strings.Join(args, " "))

	err := cmd.Execute()
	if err != nil {
		slog.Error("error while executing oras", "cmd", strings.Join(args, " "), "err", err)
		return err
	}

	return nil
}
