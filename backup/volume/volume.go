package volume

import (
	"net"
	"net/http"
	"bufio"
	"io/ioutil"
	"encoding/json"

	"fmt"
)

type (
	List interface {
		Read()
		Extract()
	}

	volume struct {
		socket socket
		version string
		active *docker
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
			}
		}
	}
)

func New() volume {
	v := volume{socket: socket{network: "unix", address: "/var/run/docker.sock"}, version: "v1.18", active: &docker{}}
	return v
}

func (o *volume) Read() error {
	conn,err := net.Dial(o.socket.network, o.socket.address)
	if err != nil {
		return err
	}
	request, _ := http.NewRequest("GET", "http://" + o.version + "/volumes?filters={%22dangling%22:[%22false%22]}", nil)
	request.Write(conn)
	response, _ := http.ReadResponse(bufio.NewReader(conn), request)
	buffer, _ := ioutil.ReadAll(response.Body)
	json.Unmarshal([]byte(string(buffer)), o.active)
	return nil
}

func (o *volume) Extract()  {
	
}

func (o *volume) PrintAll() {
	for _, s := range o.active.Volumes {
		fmt.Println(s.Name)
	}
}