package client

import (
	"net/url"
	"testing"
)

func TestClient(t *testing.T) {
	demo_server, _ := url.Parse("https://demo.navidrome.org")
	c := SubsonicClient{
		Server:   *demo_server,
		Username: "demo",
		Password: "demo",
	}

	if err := c.Ping(); err != nil {
		t.Fatal(err)
	}
	if entries, err := c.Indexes(); err != nil || entries[0].Name != "2 Mello" {
		t.Fatal(err)
	} else if childs, err := c.Dir(entries[0].ID); err != nil || childs[0].Name != "Chrono Jigga" {
		t.Fatal(err)
	} else {
		if song_list, err := c.Dir(childs[0].ID); err != nil {
			t.Fatal(err)
		} else if _, err := c.Stream(song_list[0].ID); err != nil {
			t.Fatal(err)
		} else if _, err := c.Cover(*song_list[0].CoverId); err != nil {
			t.Fatal(err)
		}

	}
}
