package backup

import (
	"net"
	"fmt"
	"time"
	"context"
	"errors"
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
		state volumeState
	}

	volumeState uint
)

const (
	using volumeState = iota
	unuse 
	inactive
)

// New make backup target
func New(network, address, version, dest string) Target {
	return Target{socket: socket{network: network,address: address}, version: version, dest: dest}
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
	activeVolume := make(map[string]*volume)
	for {
		for _, s := range activeVolume {
			s.state = unuse
		}
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
				if val, ok := activeVolume[s.Name]; ok {
					val.state = using
					if val.interval != *s.Labels.Interval {
						fmt.Println("Update " + s.Name + " volume")
						val.cancel()
						val.interval = *s.Labels.Interval
						cctx, cc := context.WithCancel(ctx)
						val.cancel = cc
						go schedule(cctx, *t, val.interval, s.Name)
					}
				}else{
					fmt.Println("Detect " + s.Name + "volume")
					cctx, cc := context.WithCancel(ctx)
					activeVolume[s.Name] = &volume{interval: *s.Labels.Interval, cancel: cc, state: using}
					go schedule(cctx, *t, *s.Labels.Interval, s.Name)
				}
			}
			for _, s := range activeVolume {
				if s.state == unuse {
					s.cancel()
					s.state = inactive
				}
			}
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
				fmt.Println("Failed " + name + "backup")
				return
			}
		case <- ctx.Done():
			fmt.Println("Cancel schedule of " + name + " backup")
			return
		}
	}
}

func backup(t Target, vname string) error {
	conn, err := net.Dial(t.socket.network, t.socket.address)
	if err != nil {
		return err
	}
	defer conn.Close()
	cname, err := createContainer(conn, t.version, vname, t.dest)
	if err != nil {
		return err
	}
	for {
		if err := startContainer(conn, t.version, cname, vname); err != nil {
			var derr *DockerAPIError
			if errors.As(err, &derr) {
				if derr.Code == 404 {
					time.Sleep(1 * time.Second)
					continue
				}
			}
			return err
		}
		break
	}
	if err := waitStopContainer(conn, t.version, cname, vname); err != nil {
		return err
	}
	if err := deleteContainer(conn, t.version, cname, vname); err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
