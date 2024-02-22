package nboci

import (
	"log/slog"
	"os/exec"
)

func Command(cmd string, args ...string) error {
	binary, err := exec.LookPath(cmd)
	if err != nil {
		ExitWithErrorf("cannot find '%s' on path, install zstd compressor", err, binary)
	}

	slog.Debug("executing", "bin", binary, "args", args)
	c := exec.Command(binary, args...)
	c.Stdout = SlogWriter{Logger: slog.Default(), Level: slog.LevelDebug}
	c.Stderr = SlogWriter{Logger: slog.Default(), Level: slog.LevelWarn}

	return c.Run()
}
