package main

import (
	"context"
	"github.com/negineri/backpacker/backup"
)

func main() {
	v := backup.New("unix", "/var/run/docker.sock")
	ctx := context.Background()
	v.Monitor(ctx)
}
