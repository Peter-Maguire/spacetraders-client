package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type IncomingResponse struct {
	Response *http.Response
	Error    error
}

type OutgoingRequest struct {
	Req           *http.Request
	ReturnChannel chan IncomingResponse
}

var requestBuffer = make([]*OutgoingRequest, 0)

var Waiting = 0

var IsRunningRequests = false

var token = fmt.Sprintf("Bearer %s", os.Getenv("TOKEN"))

func Request[T any](method string, path string, body any) (*T, *HttpError) {
	// We can't set this to bytes.Buffer type because net/http assumes data of that type will not be nil
	var buf io.Reader = nil
	if body != nil {
		data, err := json.Marshal(body)
		//fmt.Println(string(data))
		if err != nil {
			return nil, InternalError(err)
		}
		buf = bytes.NewBuffer(data)
	}
	req, err := http.NewRequest(method, fmt.Sprintf("https://api.spacetraders.io/v2/%s", path), buf)
	if err != nil {
		return nil, InternalError(err)
	}
	if body != nil {
		req.Header.Set("content-type", "application/json")
	}
	req.Header.Add("authorization", token)
	returnChan := make(chan IncomingResponse)
	Waiting++
	requestBuffer = append(requestBuffer, &OutgoingRequest{
		Req:           req,
		ReturnChannel: returnChan,
	})
	if !IsRunningRequests {
		requestLoop()
	}
	resp := <-returnChan
	Waiting--
	if resp.Error != nil {
		return nil, InternalError(resp.Error)
	}
	data, err := io.ReadAll(resp.Response.Body)
	//fmt.Println(string(data))
	if err != nil {
		return nil, InternalError(err)
	}
	output := &HttpResponse[T]{}
	err = json.Unmarshal(data, output)
	if output.Error != nil {
		return output.Data, output.Error
	}
	if err != nil {
		return output.Data, InternalError(err)
	}
	return output.Data, nil
}

func doRequests() bool {
	if len(requestBuffer) == 0 {
		return false
	}
	or := requestBuffer[0]
	requestBuffer = requestBuffer[1:]
	res, err := http.DefaultClient.Do(or.Req)
	or.ReturnChannel <- IncomingResponse{
		Response: res,
		Error:    err,
	}
	return true
}

func requestLoop() {
	IsRunningRequests = true
	go func() {
		for {
			if !doRequests() {
				IsRunningRequests = false
				break
			}
			<-time.Tick(time.Second * 1)
		}
	}()
}
