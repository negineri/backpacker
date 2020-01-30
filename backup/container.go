package backup

import (
	"net"
	"net/http"
	"bufio"
	"bytes"
	"fmt"
)

func createContainer(conn net.Conn, version, name, dest string) error {
/*	jsonStr := `{
		"Image": "alpine:latest",
		"HostConfig": {"Binds": ["` + name + `:/mnt/src", "` + t.dest + `:/mnt/dest"]},
		"Cmd": ["/bin/sh", "-c", "mkdir -p ` + name + ` && cp -av /mnt/src/* /mnt/dest/` + name + `"]
		}`
		*/
	jsonStr := `{
		"Image": "alpine:latest",
		"HostConfig": {"Binds": ["` + name + `:/mnt/src", "` + dest + `:/mnt/dest"]},
		"Cmd": ["/bin/sh", "-c", "while sleep 1000; do :; done"]
		}`
	request, err := http.NewRequest("POST", "http://" + version + "/containers/create", bytes.NewBuffer([]byte(jsonStr)))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	if err := request.Write(conn); err != nil {
		return err
	}
	response, err := http.ReadResponse(bufio.NewReader(conn), request)
	if err != nil {
		return err
	}
	switch response.StatusCode {
	case 201: //no error
		fmt.Println("Start " + name + " backup")
	case 400: //bad parameter
	case 404: //no such container
	case 406: //impossible to attach (container not running)
	case 409: //conflict
	case 500: //server error
	}
	return nil
}

func waitContainer(conn net.Conn, version, name string) error {
	request, err := http.NewRequest("POST", "http://" + version + "/containers/" + name + "/wait", nil)
	if err != nil {
		return err
	}
	if err := request.Write(conn); err != nil {
		return err
	}
	response, err := http.ReadResponse(bufio.NewReader(conn), request)
	if err != nil {
		return err
	}
	switch response.StatusCode {
	case 200: //no error
		fmt.Println("End " + name + " backup")
	case 404: //no such container
	case 500: //server error
	}
	return nil
}

func deleteContainer(conn net.Conn, version, name string) error {
	request, err := http.NewRequest("DELETE", "http://" + version + "/containers/" + name, nil)
	if err != nil {
		return err
	}
	if err := request.Write(conn); err != nil {
		return err
	}
	response, err := http.ReadResponse(bufio.NewReader(conn), request)
	if err != nil {
		return err
	}
	switch response.StatusCode {
	case 204: //no error
		fmt.Println("DELETE " + name + " backup container")
	case 400: //bad parameter
	case 404: //no such container
	case 409: //conflict
	case 500: //server error
	}
	return nil
}