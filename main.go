package main

import (
	"crypto/ed25519"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
)

var (
	//go:embed static/index.html
	index []byte
)

type CustomJWT struct {
	auth AuthInfo
	jwt.RegisteredClaims
}

type AuthInfo struct {
	Server   *string `json:"server"`
	Username *string `json:"username"`
	Password *string `json:"password"`
}

func main() {
	_, priv_key, _ := ed25519.GenerateKey(nil)

	var listen_addr string
	if len(os.Args) > 1 {
		listen_addr = os.Args[1]
	} else {
		listen_addr = "localhost:8080"
	}

	r := chi.NewRouter()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "index/html")
		w.Header().Add("Content-Length", fmt.Sprintf("%d", len(index)))
		w.Write(index)
	})

	api_r := chi.NewRouter()
	api_r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
		d := json.NewDecoder(r.Body)
		var auth AuthInfo
		if err := d.Decode(&auth); err != nil || auth.Server == nil || auth.Username == nil || auth.Password == nil {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			server_url, err := url.Parse(*auth.Server)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			client := SubsonicClient{
				server:   *server_url,
				username: *auth.Username,
				password: *auth.Password,
			}
			if err := client.ping(); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			token := jwt.NewWithClaims(&jwt.SigningMethodEd25519{}, CustomJWT{
				auth,
				jwt.RegisteredClaims{
					Issuer: "subsonic-webui",
				},
			})
			if digest, err := token.SignedString(priv_key); err != nil {
				fmt.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				digest_bytes := []byte(digest)
				w.Header().Add("Content-Type", "text/plain")
				w.Header().Add("Content-Length", fmt.Sprintf("%d", len(digest_bytes)))
				w.Write(digest_bytes)
			}
		}
	})
	r.Mount("/api/", http.StripPrefix("/api/", api_r))

	if err := http.ListenAndServe(listen_addr, r); err != nil {
		fmt.Println(err)
	}
}
