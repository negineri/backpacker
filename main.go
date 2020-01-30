package main

import (
	"context"
	"github.com/negineri/backpacker/backup"
	"os"
)

func main() {
	dest := os.Getenv("BACKPACKER_DEST")
	if dest == "" {
		dest = "/mnt/hdd1/backup/docker"
	}
	v := backup.New("unix", "/var/run/docker.sock", "v1.24", dest)
	ctx := context.Background()
	v.Monitor(ctx)
}
