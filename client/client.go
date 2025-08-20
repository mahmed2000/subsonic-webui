package client

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
	Server   url.URL
	Username string
	Password string
}

type param struct {
	query string
	value string
}

var rand_src = rand.New(rand.NewSource(time.Now().Unix()))

var (
	ErrBadStatus = fmt.Errorf("Remote server returned non-200 OK response")
	ErrNotMedia  = fmt.Errorf("Media endpoint returned non-media content")
)

// helper function, just checks if the response has the error field set, which means the response is invalid for
// whatever is trying to call it. Also ensures any response is valid json, and fits the subsonic response schema
func validate_api(r *http.Response) (*subsonicResponse, error) {
	d := json.NewDecoder(r.Body)
	var subsonic_r response
	if err := d.Decode(&subsonic_r); err != nil {
		return nil, err
	}
	if subsonic_r.SubsonicResponse.Error != nil {
		return nil, fmt.Errorf("%s", subsonic_r.SubsonicResponse.Error.Message)
	}
	return &subsonic_r.SubsonicResponse, nil
}

// Test connectivity and credentials
func (c *SubsonicClient) Ping() error {
	resp, err := c.get("/ping", nil)
	if err != nil {
		return err
	}
	if _, err := validate_api(resp); err != nil {
		return err
	}
	return nil
}

// main func, appends needed params from https://opensubsonic.netlify.app/docs/api-reference/
// performs the actual request and returns the response object, or an error
func (c *SubsonicClient) get(path string, extra_params []param) (*http.Response, error) {
	req_url := c.Server
	req_url.Path, _ = url.JoinPath("rest", path)

	q := make(url.Values)
	q.Add("f", "json")
	q.Add("v", "1.16.0")
	q.Add("c", "subsonic-webui")
	q.Add("u", c.Username)
	salt := make([]byte, 8)
	rand_src.Read(salt)
	salt_hex := fmt.Sprintf("%x", salt)
	q.Add("s", salt_hex)
	token := md5.Sum([]byte(c.Password + salt_hex))
	q.Add("t", fmt.Sprintf("%x", token))
	for _, p := range extra_params {
		q.Add(p.query, p.value)
	}
	req_url.RawQuery = q.Encode()

	resp, err := http.Get(req_url.String())
	if err != nil {
		return nil, err
	} else if resp.StatusCode != 200 {
		return nil, ErrBadStatus
	}
	return resp, nil
}

// fetches the index content, and transforms them into standard "Entry" objects
func (c *SubsonicClient) Indexes() ([]Entry, error) {
	resp, err := c.get("/getIndexes", nil)
	if err != nil {
		return nil, err
	}
	if index_resp, err := validate_api(resp); err != nil {
		return nil, err
	} else {
		entries := make([]Entry, 0)
		for _, i := range index_resp.Indexes.Index {
			// All indexes are necessarily Dirs...
			for _, j := range i.Artist {
				entries = append(entries, Entry{
					ID:      j.ID,
					IsDir:   true,
					Name:    j.Name,
					Album:   nil,
					CoverId: nil,
				})
			}
		}
		for _, i := range index_resp.Indexes.Child {
			// ... but not all child's are files
			entries = append(entries, Entry{
				IsDir:   i.IsDir,
				ID:      i.ID,
				Name:    i.Title,
				Album:   i.Album,
				CoverId: i.CoverId,
			})
		}

		return entries, nil
	}
}

// /getMusicDirectory call
// fetches dir content as a bunch of "Entry"s
func (c *SubsonicClient) Dir(id string) ([]Entry, error) {
	resp, err := c.get("/getMusicDirectory", []param{{
		query: "id",
		value: id,
	}})
	if err != nil {
		return nil, err
	}
	if dir_resp, err := validate_api(resp); err != nil {
		return nil, err
	} else {
		entries := make([]Entry, 0)
		for _, i := range dir_resp.Directory.Child {
			entries = append(entries, Entry{
				ID:      i.ID,
				IsDir:   i.IsDir,
				Name:    i.Title,
				Album:   i.Album,
				CoverId: i.CoverId,
			})
		}

		return entries, nil
	}
}

// "/stream"s a given song by its id.
func (c *SubsonicClient) Stream(id string) (*http.Response, error) {
	resp, err := c.get("/stream", []param{{
		query: "id",
		value: id,
	}})
	if err != nil {
		return nil, err
	}
	if resp.Header.Get("Content-Type") == "application/json" {
		return nil, ErrNotMedia
	}
	return resp, nil
}

// fetches cover art content by its url
func (c *SubsonicClient) Cover(id string) (*http.Response, error) {
	resp, err := c.get("/getCoverArt", []param{{
		query: "id",
		value: id,
	}})
	if err != nil {
		return nil, err
	}
	if resp.Header.Get("Content-Type") == "application/json" {
		return nil, ErrNotMedia
	}
	return resp, nil
}

// decoding and encoding structs.
// see https://opensubsonic.netlify.app/docs/

type response struct {
	SubsonicResponse subsonicResponse `json:"subsonic-response"`
}

type subsonicResponse struct {
	Status    string         `json:"status"`
	Error     *subsonicError `json:"error"`
	Indexes   *indexes       `json:"indexes"`
	Directory *directory     `json:"directory"`
}

type subsonicError struct {
	Message string `json:"message"`
}

type indexes struct {
	Index []index `json:"index"`
	Child []child `json:"child"`
}

type index struct {
	Artist []artist `json:"artist"`
}

type artist struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type child struct {
	ID      string  `json:"id"`
	IsDir   bool    `json:"isDir"`
	Title   string  `json:"title"`
	Album   *string `json:"album"`
	CoverId *string `json:"coverArt"`
}

type Entry struct {
	ID      string  `json:"id"`
	IsDir   bool    `json:"isDir"`
	Name    string  `json:"name"`
	Album   *string `json:"album,omitempty"`
	CoverId *string `json:"coverId,omitempty"`
}

type directory struct {
	Child []child `json:"child"`
}
