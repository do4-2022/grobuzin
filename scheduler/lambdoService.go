package scheduler

import (
	"fmt"
	"log"


	"github.com/do4-2022/grobuzin/database"
	"github.com/do4-2022/grobuzin/objectStorage"

	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

const (
	IDTimestampSeparator = "_"
)

type LambdoService struct {
	URL 		string
	BucketURL 	string
}

type RootFSInfo struct {
	ID 			string `json:"id"`
	Location 	string `json:"location"`
}

type LambdoSpawnRequest struct {
	// URL to the rootfs of the function
	Rootfs 			RootFSInfo	`json:"rootfs"`
	// Ports that the virtual machine needs to be exposed
	// right now we only support one port
	RequestedPorts	[]uint16	`json:"requestedPorts"`
}

type LambdoSpawnResponse struct {
	ID		string		`json:"id"`
	// Ports mapped by lambdo, leading to the requested ports
	// this a tuple under the form [host_port, vm_port]
	// for now we only support one port
	Ports	[][2]uint16 	`json:"port_mapping"` 
}

func (service *LambdoService) SpawnVM(function database.Function) (data LambdoSpawnResponse, err error) {
	var res *http.Response
	defer func() {
		if res != nil {
			res.Body.Close()
		}
	}()

	log.Println("Spawning VM for function", function.ID.String())
	
	RootfsURL := fmt.Sprint(service.BucketURL, "/", function.ID.String(), objectStorage.RooFSFile)

	body, err := json.Marshal(&LambdoSpawnRequest{
		Rootfs: RootFSInfo{
			ID: fmt.Sprint(function.ID.String(), IDTimestampSeparator, function.BuildTimestamp),
			Location: RootfsURL,
		},
		RequestedPorts: []uint16{8080}, // for now there's only a gin gonic instance for the agent is serving on 8080
	})	

	if err != nil {
		return 
	}

	res, err = http.Post(
		fmt.Sprint(service.URL, "/spawn"), 
		"application/json", 
		bytes.NewReader(body),
	)

	if err != nil {
		return
	}

	bytes, err := io.ReadAll(res.Body)

	if err != nil {
		return 
	}
	
	err = json.Unmarshal(bytes, &data)

	return 
}

func (service *LambdoService) DeleteVM(VMID string) (err error) {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprint(service.URL, "/destroy/", VMID ), nil)
	if err != nil {
		return
	}
	_, err = http.DefaultClient.Do(req)

	return
}