package entity

import (
	"encoding/json"
	"io"
	"net/http"
)

type ServerStatus struct {
	Status       string       `json:"status"`
	Version      string       `json:"version"`
	ResetDate    string       `json:"resetDate"`
	ServerResets ServerResets `json:"serverResets"`
	Description  string       `json:"description"`
}

type ServerResets struct {
	Next      string `json:"next"`
	Frequency string `json:"frequency"`
}

func GetServerStatus() (*ServerStatus, error) {
	resp, err := http.Get("https://api.spacetraders.io/v2/")
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	serverStatus := new(ServerStatus)
	err = json.Unmarshal(data, &serverStatus)

	return serverStatus, err
}
