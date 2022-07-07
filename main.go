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

func badRequest(w http.ResponseWriter, addr string) {
	log.Println("bad request")
	w.WriteHeader(http.StatusBadRequest)
	if _, err := w.Write([]byte("Bad request")); err != nil {
		log.Printf(
			"failed to write response to '%s': %v\n",
			addr,
			err,
		)
	}
}

func unauthorized(w http.ResponseWriter, addr string) {
	log.Println("unauthorized")
	w.WriteHeader(http.StatusUnauthorized)
	if _, err := w.Write([]byte("Unauthorized")); err != nil {
		log.Printf(
			"failed to write response to '%s': %v\n",
			addr,
			err,
		)
	}
}

func runScript(name string) error {
	path := fmt.Sprintf("./scripts/%s", name)
	log.Printf("running %s", path)

	cmd := exec.Command(
		"bash",
		"-c",
		path,
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	log.Println(string(out))

	return nil
}

// DeployHandler runs some script to upgrade some service
func DeployHandler(w http.ResponseWriter, r *http.Request) {
	logRequest(r)
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")

	barePath := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(barePath, "/")

	cfg, ok := config[parts[1]]
	if !ok {
		unauthorized(w, r.RemoteAddr)
		return
	}

	values, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		badRequest(w, r.RemoteAddr)
		return
	}
	secrets, ok := values["secret"]
	if !ok || len(secrets) != 1 {
		badRequest(w, r.RemoteAddr)
		return
	}

	if cfg.Secret != secrets[0] {
		unauthorized(w, r.RemoteAddr)
		return
	}

	// sync mode
	if values.Get("sync") == "true" {
		if err = runScript(cfg.Command); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("failed to run script: %v", err)

			if _, err := w.Write([]byte("deployment failed")); err != nil {
				log.Printf("Failed to write response: %v", err)
			}

			return
		}

		log.Println("deployment succeed")
		if _, err := w.Write([]byte("deployment succeed")); err != nil {
			log.Printf("Failed to write response: %v", err)
		}

		return
	}

	// async mode
	if _, err := w.Write([]byte("deploying...")); err != nil {
		log.Printf("Failed to write response: %v", err)
	}

	go func() {
		if err = runScript(cfg.Command); err != nil {
			log.Printf("failed to run script: %v", err)
		}
	}()
}

func main() {
	readConfig()
	http.HandleFunc("/deploy/", DeployHandler)
	if err := http.ListenAndServeTLS("0.0.0.0:443", "server.crt", "server.key", nil); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
