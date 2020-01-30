package backup

import (
	"net"
	"net/http"
	"bufio"
	"io/ioutil"
	"encoding/json"
)

type (	
	docker struct{
		Volumes []struct{
			Name string
			Labels struct {
				Interval *string `json:"com.negineri.backpacker.interval"`
			}
		}
	}
)

func getVolumeList(conn net.Conn, version string) (docker, error) {
	request, err := http.NewRequest("GET", "http://" + version + "/volumes?filters={%22dangling%22:[%22false%22]}", nil)
	if err != nil {
		return docker{}, err
	}
	if err := request.Write(conn); err != nil {
		return docker{}, err
	}
	response, err := http.ReadResponse(bufio.NewReader(conn), request)
	if err != nil {
		return docker{}, err
	}
	buffer, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return docker{}, err
	}
	volumes := docker{}
	if err := json.Unmarshal([]byte(string(buffer)), &volumes); err != nil {
		return docker{}, err
	}
	return volumes, nil
}