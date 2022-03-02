package web

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestWeb(t *testing.T) {
	srvCfg := SrvConfig{
		ListenAddress: ":8080",
		ReadTimeout:   "10s",
		WriteTimeout:  "10s",
	}

	server, router, err := Setup(srvCfg)
	if err != nil {
		t.Fatalf("Server setup failed: %s", err)
	}

	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// TODO Determine how to use Run() instead of Start()
	err = server.Start()
	if err != nil {
		t.Fatalf("Server failed to start: %s", err)
	}

	req, err := http.NewRequest("GET", "http://localhost:8080/ping", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %s", err)
	}
	resp, err := http.DefaultClient.Do(req.WithContext(context.Background()))
	if err != nil {
		t.Fatalf("Failed to reach server: %s", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read http response: %s", err)
	}

	fmt.Printf("Expected pong, got %s", string(body))

	if string(body) != "pong" {
		t.Fatalf("Unexpected server response: %s", err)
	}
}
