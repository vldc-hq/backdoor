package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
)

type deploymentConfig struct {
	// authorization key string
	Secret string `json:"secret"`
	// command to run
	Command string `json:"command"`
}

var config map[string]deploymentConfig

func readConfig() {
	config = make(map[string]deploymentConfig)
	fname := os.Getenv("CONFIG_FILE")
	if fname == "" {
		fname = "config.json"
	}
	b, err := ioutil.ReadFile(fname)
	if err != nil {
		log.Fatalf("failed to read config file: %v", err)
	}
	if err := json.Unmarshal(b, &config); err != nil {
		log.Fatalf("failed to parse config file: %v", err)
	}
}

func logRequest(r *http.Request) {
	log.Printf("%s %s: %s %s\n", r.RemoteAddr, r.Referer(), r.Method, r.RequestURI)
}

func badRequest(w http.ResponseWriter) {
	log.Println("bad request")
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("Bad request"))
}

func unauthorized(w http.ResponseWriter) {
	log.Println("unauthorized")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte("Unauthorized"))
}

func runScript(name string) {
	path := fmt.Sprintf("./scripts/%s", name)
	log.Printf("running %s", path)
	cmd := exec.Command(
		"bash",
		"-c",
		path,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Println(err)
	}
	log.Println(string(out))
}

// DeployHandler runs some script to upgrade some service
func DeployHandler(w http.ResponseWriter, r *http.Request) {
	logRequest(r)
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")

	barePath := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(barePath, "/")

	cfg, ok := config[parts[1]]
	if !ok {
		unauthorized(w)
		return
	}

	values, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		badRequest(w)
		return
	}
	secrets, ok := values["secret"]
	if !ok || len(secrets) != 1 {
		badRequest(w)
		return
	}

	if cfg.Secret != secrets[0] {
		unauthorized(w)
		return
	}

	if _, err := w.Write([]byte("deploying...")); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
	go runScript(cfg.Command)
}

func main() {
	readConfig()
	http.HandleFunc("/deploy/", DeployHandler)
	if err := http.ListenAndServeTLS("0.0.0.0:8443", "server.crt", "server.key", nil); err != nil {
		log.Fatalf("failed to start server")
	}
}
