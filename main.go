package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"subsonic-webui/client"

	"github.com/go-chi/chi/v5"
)

var (
	//go:embed index.html
	index []byte
)

const (
	SERVER   = "SUBSONIC_SERVER"
	USERNAME = "SUBSONIC_USERNAME"
	PASSWORD = "SUBSONIC_PASSWORD"
)

func main() {
	// check input params
	var listen_addr string
	if len(os.Args) > 1 {
		listen_addr = os.Args[1]
	} else {
		listen_addr = "localhost:8080"
	}

	server, ok := os.LookupEnv(SERVER)
	if !ok {
		fmt.Println("Missing server in", SERVER)
		os.Exit(1)
	}
	server_url, err := url.Parse(server)
	if err != nil {
		fmt.Println(server, "not a valid server URL")
		os.Exit(1)
	}
	username, ok := os.LookupEnv(USERNAME)
	if !ok {
		fmt.Println("Missing Username in", USERNAME)
		os.Exit(1)
	}
	password, ok := os.LookupEnv(PASSWORD)
	if !ok {
		fmt.Println("Missing Username in", PASSWORD)
		os.Exit(1)
	}
	c := client.SubsonicClient{
		Server:   *server_url,
		Username: username,
		Password: password,
	}
	if err := c.Ping(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// this would ordinarily be its own class. This app is small enough to not warrant the extra effort
	r := chi.NewRouter()
	// serve the index.html file
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		w.Header().Add("Content-Length", fmt.Sprintf("%d", len(index)))
		w.Write(index)
	})

	// subrouter for all paths on /api/
	api_r := chi.NewRouter()
	api_r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		if err := c.Ping(); err != nil {
			w.WriteHeader(http.StatusBadGateway)
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	})
	api_r.Get("/list/{dir_id}", func(w http.ResponseWriter, r *http.Request) {
		var entries []client.Entry
		var err error
		// if target dir is magic NULL byte (signifies root dir/indexes)
		if dir_id := chi.URLParam(r, "dir_id"); dir_id[0] == 0x00 {
			entries, err = c.Indexes()
		} else {
			entries, err = c.Dir(dir_id)
		}

		// Outside of network errors, this should never error in normal use
		// Ideally there would be checking on the response here instead of a blanket response status
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
		} else {
			data, _ := json.Marshal(entries)
			w.Header().Add("Content-Type", "application/json")
			w.Header().Add("Content-Length", fmt.Sprintf("%d", len(data)))
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		}
	})
	api_r.Get("/stream/{song_id}", func(w http.ResponseWriter, r *http.Request) {
		if resp, err := c.Stream(chi.URLParam(r, "song_id")); err != nil {
			w.WriteHeader(http.StatusBadGateway)
		} else {
			w.Header().Add("Content-Type", resp.Header.Get("Content-Type"))
			w.Header().Add("Content-Length", resp.Header.Get("Content-Length"))
			w.WriteHeader(http.StatusOK)
			io.Copy(w, resp.Body)
		}
	})
	api_r.Get("/cover/{cover_id}", func(w http.ResponseWriter, r *http.Request) {
		if resp, err := c.Cover(chi.URLParam(r, "cover_id")); err != nil {
			w.WriteHeader(http.StatusBadGateway)
		} else {
			w.Header().Add("Content-Type", resp.Header.Get("Content-Type"))
			w.Header().Add("Content-Length", resp.Header.Get("Content-Length"))
			w.WriteHeader(http.StatusOK)
			io.Copy(w, resp.Body)
		}
	})
	r.Mount("/api/", http.StripPrefix("/api/", api_r))

	if err := http.ListenAndServe(listen_addr, r); err != nil {
		fmt.Println(err)
	}
}
