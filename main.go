package main

import (
	"context"
	"fmt"
	"github.com/negineri/backpacker/backup"
	"os"
)

func main() {
	dest := os.Getenv("BACKPACKER_DEST")
	if dest == "" {
		fmt.Println("Please set BACKPACKER_DEST")
		return
	}
	v := backup.New("unix", "/var/run/docker.sock", "v1.24", dest)
	ctx := context.Background()
	v.Monitor(ctx)
}
