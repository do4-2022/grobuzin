package scheduler

import (
	"fmt"

	"github.com/google/uuid"

	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type LambdoService struct {
	URL 		string
	MinioURL 	string
}

type LambdoSpawnRequest struct {
	// URL to the rootfs of the function
	RootfsURL 		string		`json:"rootfs"`
	// Ports that the virtual machine needs to be exposed
	// right now we only support one port
	RequestedPorts	[]uint16	`json:"requestedPorts"`
}

type LambdoSpawnResponse struct {
	ID		string		`json:"ID"`
	// Ports mapped by lambdo, leading to the requested ports
	Ports	[]uint16 	`json:"ports"`
}

func (service *LambdoService) SpawnVM(function_id uuid.UUID) (data LambdoSpawnResponse, err error) {
	var res *http.Response
	defer func() {
		if res != nil {
			res.Body.Close()
		}
	}()

	body, err := json.Marshal(&LambdoSpawnRequest{
		RootfsURL: fmt.Sprintf("%s/%s", service.MinioURL, function_id),
	})

	if err != nil {
		return 
	}

	res, err = http.Post(
		fmt.Sprintf(service.URL, "/spawn"), 
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