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
		intervalTemp string
		cancel context.CancelFunc
	}
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
		select {
		case <- ctx.Done():
			return
		case <- wait.C:
			volumes, err := getVolumeList(conn, t.version)
			if err != nil {
				return
			}
			for _, s := range volumes.Volumes {
				if _, ok := activeVolume[s.Name]; ok == false {
					activeVolume[s.Name] = &volume{}
				}
				if s.Labels.Interval != nil {
					activeVolume[s.Name].intervalTemp = *s.Labels.Interval
				}
			}
			containerVolumes, err := getContainerList(conn, t.version)
			if err != nil {
				return
			}
			for _, s := range containerVolumes {
				if s.Labels.Interval == nil {
					continue
				}
				for _, m := range s.Mounts {
					if m.Type != "volume" {
						continue
					}
					if priorityInterval(activeVolume[m.Name].intervalTemp) < priorityInterval(*s.Labels.Interval) {
						activeVolume[m.Name].intervalTemp = *s.Labels.Interval
					}
				}
			}
			for i, s := range activeVolume {
				if priorityInterval(s.intervalTemp) == 0 {
					if priorityInterval(s.interval) != 0 {
						s.cancel()
						s.interval = ""
					}
					continue
				}
				if priorityInterval(s.interval) == priorityInterval(s.intervalTemp) {
					s.intervalTemp = ""
					continue
				}
				if priorityInterval(s.interval) != 0 {
					s.cancel()
				}
				fmt.Println("Detect " + i + "volume")
				s.interval = s.intervalTemp
				s.intervalTemp = ""
				cctx, cc := context.WithCancel(ctx)
				s.cancel = cc
				go schedule(cctx, *t, s.interval, i)
			}
		}
	}
}

func priorityInterval(a string) int {
	switch a {
	case "hourly":
		return 50
	case "daily":
		return 40
	case "weekly":
		return 30
	default:
		return 0
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
