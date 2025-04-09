package client

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/dev-hack95/mini/structs"
	"github.com/dev-hack95/mini/utilities"
)

type Client struct {
	Model  string `json:"model"`
	Url    string `json:"url"`
	Role   string `json:"role"`
	Stream bool   `json:"stream"`
}

func NewClient() (*Client, error) {

	data, err := utilities.CheckFileExist("/.term-ollama/model.yaml")

	if err != nil {
		return nil, err
	}

	return &Client{
		Model:  data.Model,
		Url:    data.Url,
		Role:   data.Role,
		Stream: data.Stream,
	}, nil
}

func (c *Client) ChatOllama(message []structs.Message) (*structs.Respone, error) {
	msg := structs.Message{
		Role:    message[len(message)-1].Role,
		Content: message[len(message)-1].Content,
	}

	message = append(message, msg)

	req := structs.Request{
		Model:    c.Model,
		Stream:   c.Stream,
		Messages: message,
	}

	req.Options.Temperature = 0.8

	js, err := json.Marshal(req)

	if err != nil {
		return nil, err
	}

	client := http.Client{}
	httpReq, err := http.NewRequest(http.MethodPost, c.Url, bytes.NewReader(js))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	var ollamaResp structs.Respone
	err = json.NewDecoder(httpResp.Body).Decode(&ollamaResp)
	return &ollamaResp, err

}

// func (c *Client) ChatOllamaSearch(message []structs.Message) (*structs.Respone, error) {
// }
