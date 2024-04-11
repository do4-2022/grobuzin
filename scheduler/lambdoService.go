package scheduler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type LambdoService struct {
	URL string
}

type LambdoCodeFileItem struct {
	Filename string `json:"filename"`
	Content string `json:"content"`
}

type LambdoRunRequest struct {
	Language string `json:"language"`
	Version string `json:"version"`
	Input string `json:"input"`
	Code []LambdoCodeFileItem `json:"code"`
}

type LambdoRunResponse struct {
	Status 	int		`json:"status"`
	Stdout 	string 	`json:"stdout"`
	Stderr 	string 	`json:"stderr"`
	Port   	uint16 	`json:"port"`    ////
	Address string	`json:"address"` //// Should be added by Simon 
}

func (service *LambdoService) RunFunction(code string) (data LambdoRunResponse, err error) {
	var res *http.Response
	defer func() {
		if res != nil {
			res.Body.Close()
		}
	}()

	body, err := json.Marshal(&LambdoRunRequest{
		Language: "NODE", // TODO
		Version: "1.0.0",
		Input: "",
		Code: []LambdoCodeFileItem{
			{
				Filename: "index.js",
				Content: code,
			},
		},
	})

	if err != nil {
		return 
	}

	res, err = http.Post(service.URL, "application/json", bytes.NewReader(body))

	if err != nil {
		return
	}

	bytes, err := io.ReadAll(res.Body)

	if err != nil {
		return 
	}
	
	data = LambdoRunResponse{}
	err = json.Unmarshal(bytes, &data)

	return
}