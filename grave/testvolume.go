package main

import (
	"os"
  "bufio"
  "fmt"
  "net"
  "net/http"
	"io/ioutil"
	"encoding/json"
	"time"
    //"net/http/httputil"
)

type volume struct {
    Volumes []struct{
				Name string
				Labels string
    }
}

func main() {
	conn, err := net.Dial("unix", "/var/run/docker.sock")
	if err != nil {
		os.Exit(1)
	}
	volume := volume{}
	for index := 0; index < 10; index++ {
		request, _ := http.NewRequest("GET", "http://v1.24/volumes?filters={%22dangling%22:[%22false%22]}", nil)
		request.Write(conn)
		response, _ := http.ReadResponse(bufio.NewReader(conn), request)
		//dump, _ := httputil.DumpResponse(response, true)
		//fmt.Println(string(dump))
		buffer, _ := ioutil.ReadAll(response.Body)
		json.Unmarshal([]byte(string(buffer)), &volume)
		fmt.Println("-------")
		for _, s := range volume.Volumes{
			fmt.Println(s.Name)
		}
		time.Sleep(3 * time.Second)
	}
}
