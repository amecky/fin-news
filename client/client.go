package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/amecky/fin-news/model"
)

type NewsClient struct {
	HTTPClient *http.Client
	BaseUrl    string
}

func NewNewsClient(baseUrl string) *NewsClient {
	return &NewsClient{
		BaseUrl: baseUrl,
		HTTPClient: &http.Client{
			Timeout: time.Minute,
		},
	}
}

func (c *NewsClient) FindRecentNews(limit int) ([]model.RssNews, error) {
	var list = make([]model.RssNews, 0)
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/news?limit=%d", c.BaseUrl, limit), nil)
	if err != nil {
		return nil, err
	}

	err = c.sendRequest(req, &list)
	return list, err
}

func (c *NewsClient) DeleteItem(id int) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/news/%d", c.BaseUrl, id), nil)
	if err != nil {
		return err
	}

	return c.sendRequest(req, nil)
}

func (c *NewsClient) sendRequest(req *http.Request, v interface{}) error {
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json; charset=utf-8")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("unknown error, status code: %d", res.StatusCode)
	}
	if v != nil {
		if err = json.NewDecoder(res.Body).Decode(&v); err != nil {
			return err
		}
	}
	return nil
}
