package agent

import (
	"alexdenkk/labs/internal/lab/domain"
	"alexdenkk/labs/pkg/config"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

type adapter struct {
	config *config.AgentConfig
}

func New(cfg *config.AgentConfig) domain.AgentAdapter {
	return &adapter{
		config: cfg,
	}
}

func (adapter *adapter) Call(
	ctx context.Context, request domain.AgentRequest, accessID string,
) (domain.AgentResponse, error) {
	requestUrl := adapter.config.BaseURL + accessID + "/call"
	println(requestUrl)

	encodedRequest, _ := json.Marshal(request)

	req, _ := http.NewRequest("POST", requestUrl, bytes.NewBuffer(encodedRequest))

	req.Header.Add("Authorization", "Bearer "+adapter.config.AccessToken)
	req.Header.Add("x-proxy-source", "")
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		println(err.Error())
		return domain.AgentResponse{}, err
	}

	if resp.Status != "200 OK" {
		println(resp.Status)
		return domain.AgentResponse{}, errors.New("error calling agent api")
	}

	var agentResp domain.AgentResponse

	if err := json.NewDecoder(resp.Body).Decode(&agentResp); err != nil {
		return domain.AgentResponse{}, errors.New("error decoding agent response")
	}

	return agentResp, nil
}
