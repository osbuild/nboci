package nboci

import (
	"log/slog"
	"os/exec"
)

func Command(cmd string, args ...string) error {
	binary, err := exec.LookPath(cmd)
	if err != nil {
		FatalfErr(err, "cannot find '%s' on path, install zstd compressor", binary)
	}

	slog.Debug("executing", "bin", binary, "args", args)
	c := exec.Command(binary, args...)
	c.Stdout = OutputWriter{}
	c.Stderr = OutputWriter{}

	return c.Run()
}
