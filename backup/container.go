package backup

import (
	"net"
	"net/http"
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"encoding/json"
)

type (
	// A DockerAPIError contains httpStatusCode
	DockerAPIError struct {
		Code int
		Message string
	}

	containerJSON struct {
		Id string
	}

)

func (err *DockerAPIError) Error() string { return err.Message }

func dockerAPIErrorf(code int, message string) error {
	return &DockerAPIError{Code: code, Message: message}
}

func createContainer(conn net.Conn, version, vname, dest string) (string, error) {
	jsonStr := `{
		"Image": "alpine:latest",
		"HostConfig": {"Binds": ["` + vname + `:/mnt/src", "` + dest + `:/mnt/dest"]},
		"Cmd": ["/bin/sh", "-c", "mkdir -p /mnt/dest/` + vname + ` && cp -av /mnt/src/* /mnt/dest/` + vname + `"]
		}`
	request, err := http.NewRequest("POST", "http://" + version + "/containers/create", bytes.NewBuffer([]byte(jsonStr)))
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", "application/json")
	if err := request.Write(conn); err != nil {
		return "", err
	}
	response, err := http.ReadResponse(bufio.NewReader(conn), request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	switch response.StatusCode {
	case 201: //no error
		fmt.Println("Prepare " + vname + " backup")
		buffer, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return "", err
		}
		responseJSON := containerJSON{}
		if err := json.Unmarshal([]byte(string(buffer)), &responseJSON); err != nil {
			return "", err
		}
		return responseJSON.Id, nil
	case 400: //bad parameter
	case 404: //no such container
	case 406: //impossible to attach (container not running)
	case 409: //conflict
	case 500: //server error
	}
	return "", dockerAPIErrorf(response.StatusCode, "")
}

func startContainer(conn net.Conn, version, cname, vname string) error {
	request, err := http.NewRequest("POST", "http://" + version + "/containers/" + cname + "/start", nil)
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
	defer response.Body.Close()
	switch response.StatusCode {
	case 204: //no error
		fmt.Println("Start " + vname + " backup")
	case 304: //container already started
	case 404: //no such container
		return dockerAPIErrorf(404, "Not found " + cname)
	case 500: //server error
	}
	return nil
}

func waitStopContainer(conn net.Conn, version, cname, vname string) error {
	request, err := http.NewRequest("POST", "http://" + version + "/containers/" + cname + "/wait", nil)
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
	defer response.Body.Close()
	switch response.StatusCode {
	case 200: //no error
		fmt.Println("End " + vname + " backup")
	case 404: //no such container
	case 500: //server error
	}
	return nil
}

func deleteContainer(conn net.Conn, version, cname, vname string) error {
	request, err := http.NewRequest("DELETE", "http://" + version + "/containers/" + cname, nil)
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
	defer response.Body.Close()
	switch response.StatusCode {
	case 204: //no error
		fmt.Println("DELETE " + vname + " backup container")
	case 400: //bad parameter
	case 404: //no such container
	case 409: //conflict
	case 500: //server error
	}
	return nil
}