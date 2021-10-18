package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"sync"
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

func runScript(name string, w http.ResponseWriter) error {
	path := fmt.Sprintf("./scripts/%s", name)
	log.Printf("running %s", path)
	cmd := exec.Command(
		"bash",
		"-c",
		path,
	)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("could not get stderr pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("could not get stdout pipe: %w", err)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()

		merged := io.MultiReader(stderr, stdout)
		scanner := bufio.NewScanner(merged)

		for scanner.Scan() {
			b := scanner.Bytes()
			log.Print(b)

			if _, err = w.Write(b); err != nil {
				log.Printf("failed to read cmd: %v", err)
				return
			}
		}
	}()

	wg.Wait()
	return err
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

	if _, err := w.Write([]byte("deploying...")); err != nil {
		log.Printf("Failed to write response: %v", err)
	}

	if err := runScript(cfg.Command, w); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("deploy failed\n")) //nolint:errcheck
		return
	}

	w.Write([]byte("done\n")) //nolint:errcheck
}

func main() {
	readConfig()
	http.HandleFunc("/deploy/", DeployHandler)
	if err := http.ListenAndServeTLS("0.0.0.0:443", "server.crt", "server.key", nil); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
