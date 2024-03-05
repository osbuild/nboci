package nboci

import (
	"os"
	"path"

	"oras.land/oras-go/v2/registry/remote/credentials"
)

func configPath() string {
	hd, err := os.UserHomeDir()
	if err != nil {
		FatalErr(err, "failed to get user home directory")
	}

	return path.Join(hd, ".config", "nboci.json")
}

func NewStore() credentials.Store {
	opts := credentials.StoreOptions{AllowPlaintextPut: true}

	s, err := credentials.NewStore(configPath(), opts)
	if err != nil {
		FatalErr(err, "failed to create store")
	}
	
	return s
}
