package main

import (
	"crypto/ed25519"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"subsonic-webui/client"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
)

var (
	//go:embed static/index.html
	index []byte

	priv_key ed25519.PrivateKey
	pub_key  ed25519.PublicKey
)

func init() {
	pub_key, priv_key, _ = ed25519.GenerateKey(nil)
}

type CustomJWT struct {
	Auth AuthInfo
	jwt.RegisteredClaims
}

type AuthInfo struct {
	Server   string `json:"server"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func main() {
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
		if err := d.Decode(&auth); err != nil {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			server_url, err := url.Parse(auth.Server)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			c := client.SubsonicClient{
				Server:   *server_url,
				Username: auth.Username,
				Password: auth.Password,
			}
			if err := c.Ping(); err != nil {
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
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				digest_bytes := []byte(digest)
				w.Header().Add("Content-Type", "text/plain")
				w.Header().Add("Content-Length", fmt.Sprintf("%d", len(digest_bytes)))
				w.Write(digest_bytes)
			}
		}
	})
	api_r.Get("/list/{dir_id}", func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if auth, err := validate_jwt(token); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			server_url, _ := url.Parse(auth.Server)
			c := client.SubsonicClient{
				Server:   *server_url,
				Username: auth.Username,
				Password: auth.Password,
			}
			var entries []client.Entry
			var err error
			if dir_id := chi.URLParam(r, "dir_id"); dir_id[0] == 0x00 {
				entries, err = c.Indexes()
			} else {
				entries, err = c.Dir(dir_id)
			}

			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
			} else {
				data, _ := json.Marshal(entries)
				w.Header().Add("Content-Type", "application/json")
				w.Header().Add("Content-Length", fmt.Sprintf("%d", len(data)))
				w.WriteHeader(http.StatusOK)
				w.Write(data)
			}
		}
	})
	r.Mount("/api/", http.StripPrefix("/api/", api_r))

	if err := http.ListenAndServe(listen_addr, r); err != nil {
		fmt.Println(err)
	}
}

func validate_jwt(token string) (*AuthInfo, error) {
	decoded, err := jwt.ParseWithClaims(token, &CustomJWT{}, func(t *jwt.Token) (any, error) {
		return pub_key, nil
	})
	if err != nil {
		return nil, err
	}
	if info, ok := decoded.Claims.(*CustomJWT); !ok {
		return nil, fmt.Errorf("Somehow signed an unnknown set of claims?")
	} else {
		return &info.Auth, nil
	}
}
