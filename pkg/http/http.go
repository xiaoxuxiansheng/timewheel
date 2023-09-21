package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Client struct {
	core *http.Client
}

func NewClient() *Client {
	return &Client{
		core: http.DefaultClient,
	}
}

func (c *Client) JSONGet(ctx context.Context, url string, header, params map[string]string, resp interface{}) error {
	return c.JSONDo(ctx, http.MethodGet, getCompleteURL(url, params), header, nil, resp)
}

func (c *Client) JSONPost(ctx context.Context, url string, header map[string]string, req, resp interface{}) error {
	return c.JSONDo(ctx, http.MethodPost, url, header, req, resp)
}

func (c *Client) JSONDo(ctx context.Context, method, url string, header map[string]string, req, resp interface{}) error {
	var reqReader io.Reader
	if req != nil {
		body, _ := json.Marshal(req)
		reqReader = bytes.NewReader(body)
	}

	request, err := http.NewRequestWithContext(ctx, method, url, reqReader)
	if err != nil {
		return err
	}

	if request.Header == nil {
		request.Header = make(http.Header)
	}
	for k, v := range header {
		request.Header.Add(k, v)
	}
	request.Header.Add("Content-Type", "application/json")

	response, err := c.core.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid status: %d", response.StatusCode)
	}

	if resp == nil {
		return nil
	}

	respBody, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(respBody, resp)
}

func getCompleteURL(origin string, params map[string]string) string {
	if len(params) == 0 {
		return origin
	}
	values := url.Values{}
	for k, v := range params {
		values.Add(k, v)
	}

	queriesStr, _ := url.QueryUnescape(values.Encode())
	return fmt.Sprintf("%s?%s", origin, queriesStr)
}
