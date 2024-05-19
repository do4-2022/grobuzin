package scheduler

import (
	"fmt"
	"log"

	"github.com/do4-2022/grobuzin/objectStorage"
	"github.com/google/uuid"

	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type LambdoService struct {
	URL 		string
	BucketURL 	string
}

type LambdoSpawnRequest struct {
	// URL to the rootfs of the function
	RootfsURL 		string		`json:"rootfs"`
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

func (service *LambdoService) SpawnVM(function_id uuid.UUID) (data LambdoSpawnResponse, err error) {
	var res *http.Response
	defer func() {
		if res != nil {
			res.Body.Close()
		}
	}()

	log.Println("Spawning VM for function", function_id)
	
	RootfsURL := fmt.Sprint(service.BucketURL, function_id.String(), objectStorage.RooFSFile)

	body, err := json.Marshal(&LambdoSpawnRequest{
		RootfsURL: RootfsURL,
		RequestedPorts: []uint16{8080}, // for now only a gin gonic instance for the agent is serving on 8080
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
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf(service.URL, "/destroy/", VMID ), nil)
	if err != nil {
		return
	}
	_, err = http.DefaultClient.Do(req)

	return
}