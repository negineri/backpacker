package backup

import (
	"net"
	"net/http"
	"bufio"
	"io/ioutil"
	"encoding/json"
	"time"
	"context"
	"fmt"
)

type (
	Driver interface {
		Monitor()
	}

	Target struct {
//		volumeSet volume.List
		socket socket
		version string
	}

	socket struct {
		network string
		address string
	}
	
	docker struct{
		Volumes []struct{
			Name string
			Labels []struct {
				Cron *string `json:"com.negineri.backpacker.cron"`
				Interval *string `json:"com.negineri.backpacker.interval"`
			}
		}
	}
)

func New(network string, address string) Target {
	return Target{socket: socket{network: network,address: address}, version: "v1.18"}
}

func (t *Target) Monitor(ctx context.Context) {
	conn,err := net.Dial(t.socket.network, t.socket.address)
	if err != nil {
		return
	}
	defer conn.Close()
	wait := time.NewTicker(10 * time.Second)
	defer wait.Stop()
	active := make(map[string]bool)
	quit := make(chan string, 10)
	defer close(quit)
	for {
		select {
		case <- ctx.Done():
			return
		case <- wait.C:
			request, _ := http.NewRequest("GET", "http://" + t.version + "/volumes?filters={%22dangling%22:[%22false%22]}", nil)
			request.Write(conn)
			response, _ := http.ReadResponse(bufio.NewReader(conn), request)
			buffer, _ := ioutil.ReadAll(response.Body)
			volumes := docker{}
			json.Unmarshal([]byte(string(buffer)), &volumes)
			for _, s := range volumes.Volumes {
				if _, ok := active[s.Name]; ok == false {
					fmt.Println(s.Name)
					active[s.Name] = true
					cctx, cc := context.WithCancel(ctx)
					go t.volume(cctx, cc, quit, s.Name)
				}
			}
		case n := <- quit:
			delete(active, n)
		}
	}
}

func (t *Target)volume(ctx context.Context, cancel context.CancelFunc, quit chan<- string, name string)  {
	conn,err := net.Dial(t.socket.network, t.socket.address)
	if err != nil {
		return
	}
	defer conn.Close()
	wait := time.NewTicker(10 * time.Second)
	defer wait.Stop()
	for {
		select {
		case <- ctx.Done():
			quit <- name
			return
		case <- wait.C:
			request, _ := http.NewRequest("GET", "http://" + t.version + "/volumes/" + name, nil)
			request.Write(conn)
			response, _ := http.ReadResponse(bufio.NewReader(conn), request)
			buffer, _ := ioutil.ReadAll(response.Body)
			volumes := docker{}
			json.Unmarshal([]byte(string(buffer)), &volumes)
			for _, s := range volumes.Volumes {
				if _, ok := active[s.Name]; ok == false {
					fmt.Println(s.Name)
					active[s.Name] = true
					cctx, cc := context.WithCancel(ctx)
					go volume(cctx, cc, quit, s.Name, t.socket)
				}
			}
		}
	}
}