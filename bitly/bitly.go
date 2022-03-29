// Package bitly is a Go package that provides a client for the bitly API v4
package bitly

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

const apiURL = "https://api-ssl.bitly.com/v4"

type Bitly struct {
	token  string
	client *http.Client
}

type ShortenRequest struct {
	LongURL string `json:"long_url"`
}

type ShortenResponse struct {
	CreatedAt      string        `json:"created_at"`
	ID             string        `json:"id"`
	Link           string        `json:"link"`
	CustomBitlinks []interface{} `json:"custom_bitlinks"`
	LongURL        string        `json:"long_url"`
	Archived       bool          `json:"archived"`
	Tags           []interface{} `json:"tags"`
	Deeplinks      []interface{} `json:"deeplinks"`
	References     struct {
		Group string `json:"group"`
	} `json:"references"`
}

func New(token string) *Bitly {
	client := &http.Client{}
	return &Bitly{
		token:  token,
		client: client,
	}
}

// Shorten takes a long URL and returns a short URL
func (b *Bitly) Shorten(link string) (string, error) {
	var sres ShortenResponse
	jsonBody, err := json.Marshal(ShortenRequest{
		LongURL: link,
	})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", apiURL+"/shorten", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", "Bearer "+b.token)
	req.Header.Add("Content-Type", "application/json")
	resp, err := b.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(body, &sres)
	if err != nil {
		return "", err
	}

	return sres.Link, nil
}
