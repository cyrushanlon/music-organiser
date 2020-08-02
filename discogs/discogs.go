package discogs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

type response struct {
	Results []*Result `json:"results,omitempty"`
}

// Result describes a part of search result
type Result struct {
	Style      []string `json:"style,omitempty"`
	Thumb      string   `json:"thumb,omitempty"`
	CoverImage string   `json:"cover_image,omitempty"`
	Title      string   `json:"title,omitempty"`
	Country    string   `json:"country,omitempty"`
	Format     []string `json:"format,omitempty"`
	URI        string   `json:"uri,omitempty"`
	// Community   Community `json:"community,omitempty"`
	Label       []string `json:"label,omitempty"`
	Catno       string   `json:"catno,omitempty"`
	Year        string   `json:"year,omitempty"`
	Genre       []string `json:"genre,omitempty"`
	ResourceURL string   `json:"resource_url,omitempty"`
	Type        string   `json:"type,omitempty"`
	ID          int      `json:"id,omitempty"`
	MasterID    int      `json:"master_id,omitempty"`
}

type Client interface {
	Search(string) ([]*Result, error)
}

type client struct {
	c   *http.Client
	url string

	auth      string
	userAgent string
}

func New(key string, secret string, userAgent string) Client {
	return &client{
		c:         &http.Client{},
		url:       "https://api.discogs.com/",
		auth:      "&key=" + key + "&secret=" + secret,
		userAgent: userAgent,
	}
}

func (c *client) Search(q string) ([]*Result, error) {
	req, err := http.NewRequest("GET", c.url+"database/search?page=0&per_page=1&q="+q+c.auth, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.c.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	i, _ := strconv.Atoi(resp.Header.Get("X-Discogs-Ratelimit-Remaining"))

	if i == 0 {

	} else if i < 5 {
		fmt.Printf("very low requests remaining(%d), sleeping for 10s\n", i)
		time.Sleep(time.Second * 10)
	} else if i < 10 {
		fmt.Printf("low requests remaining(%d), sleeping for 1s\n", i)
		time.Sleep(time.Second)
	}

	out := &response{}
	err = json.Unmarshal(body, out)
	if err != nil {
		return nil, err
	}

	return out.Results, nil
}
