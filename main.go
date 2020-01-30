package main

import (
	"context"
	"github.com/negineri/backpacker/backup"
)

func main() {
	v := backup.New("unix", "/var/run/docker.sock", "v1.24", "/mnt/hdd1/backup/docker")
	ctx := context.Background()
	v.Monitor(ctx)
}
