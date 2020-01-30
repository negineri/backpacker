package backup

import (
	"net"
	"fmt"
	"time"
	"context"
)

type (
	// Driver is object of backup target
	Driver interface {
		Monitor()
	}

	// A Target contains backup target info
	Target struct {
		socket socket
		version string
		dest string
	}

	socket struct {
		network string
		address string
	}
	
	volume struct{
		interval string
		cancel context.CancelFunc
	}
)

// New make backup target
func New(network string, address string) Target {
	return Target{socket: socket{network: network,address: address}, version: "v1.24", dest: "/mnt/hdd1/backup"}
}

// Monitor monitor docker volume list
func (t *Target) Monitor(ctx context.Context) {
	conn,err := net.Dial(t.socket.network, t.socket.address)
	if err != nil {
		return
	}
	defer conn.Close()
	wait := time.NewTicker(1 * time.Second)
	defer wait.Stop()
	active := make(map[string]*volume)
	quit := make(chan string, 10)
	defer close(quit)
	for {
		select {
		case <- ctx.Done():
			return
		case <- wait.C:
			volumes, err := getVolumeList(conn, t.version)
			if err != nil {
				return
			}
			for _, s := range volumes.Volumes {
				if s.Labels.Interval == nil {
					continue
				}
				if val, ok := active[s.Name]; ok {
					if val.interval != *s.Labels.Interval {
						val.cancel()
						val.interval = *s.Labels.Interval
						cctx, cc := context.WithCancel(ctx)
						val.cancel = cc
						go schedule(cctx, *t, val.interval, s.Name)
					}
				}else{
					fmt.Println(s.Name)
					cctx, cc := context.WithCancel(ctx)
					active[s.Name] = &volume{interval: *s.Labels.Interval, cancel: cc}
					go schedule(cctx, *t, *s.Labels.Interval, s.Name)
				}
			}
		case n := <- quit:
			delete(active, n)
		}
	}
}

func schedule(ctx context.Context, t Target, interval string, name string) {
	var ticker *time.Ticker
	switch interval {
	case "hourly":
		ticker = time.NewTicker(1 * time.Hour)
	case "daily":
		ticker = time.NewTicker(24 * time.Hour)
	case "weekly":
		ticker = time.NewTicker(168 * time.Hour)
	default:
		return
	}
	defer ticker.Stop()
	if err := backup(t, name); err != nil {
		return
	}
	for {
		select {
		case <- ticker.C:
			if err := backup(t, name); err != nil {
				return
			}
		case <-ctx.Done():
		}
	}
}

func backup(t Target, name string) error {
	conn,err := net.Dial(t.socket.network, t.socket.address)
	if err != nil {
		return err
	}
	defer conn.Close()
	if err := createContainer(conn, t.version, name, t.dest); err != nil {
		return err
	}
	if err := waitContainer(conn, t.version, name); err != nil {
		return err
	}
	if err := deleteContainer(conn, t.version, name); err != nil {
		return err
	}
	return nil
}
