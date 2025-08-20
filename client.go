package main

import (
	"crypto/md5"

	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

type SubsonicClient struct {
	server   url.URL
	username string
	password string
}

type param struct {
	query string
	value string
}

var rand_src = rand.New(rand.NewSource(time.Now().Unix()))

var ErrBadStatus = fmt.Errorf("Remote server returned non-200 OK response")

func (c *SubsonicClient) ping() error {
	resp, err := c.get("/ping", nil)
	if err != nil {
		return err
	}
	d := json.NewDecoder(resp.Body)
	var subsonic_r response
	if err := d.Decode(&subsonic_r); err != nil {
		return err
	}

	if subsonic_r.SubsonicResponse.Error != nil {
		fmt.Println(subsonic_r.SubsonicResponse.Error.Message)
		return fmt.Errorf("%s", subsonic_r.SubsonicResponse.Error.Message)
	}
	return nil
}

func (c *SubsonicClient) get(path string, extra_params []param) (*http.Response, error) {
	req_url := c.server
	req_url.Path, _ = url.JoinPath("rest", path)
	q := make(url.Values)
	q.Add("f", "json")
	q.Add("v", "1.16.0")
	q.Add("c", "subsonic-webui")
	q.Add("u", c.username)
	salt := make([]byte, 8)
	rand_src.Read(salt)
	salt_hex := fmt.Sprintf("%x", salt)
	q.Add("s", salt_hex)
	token := md5.Sum([]byte(c.password + salt_hex))
	q.Add("t", fmt.Sprintf("%x", token))
	req_url.RawQuery = q.Encode()

	resp, err := http.Get(req_url.String())
	if err != nil {
		return nil, err
	} else if resp.StatusCode != 200 {
		return nil, ErrBadStatus
	}
	return resp, nil
}

type response struct {
	SubsonicResponse subsonicResponse `json:"subsonic-response"`
}

type subsonicResponse struct {
	Status string         `json:"status"`
	Error  *subsonicError `json:"error"`
}

type subsonicError struct {
	Message string `json:"message"`
}
