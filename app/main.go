package main

import (
	"fmt"
	"net/http"
	"os"

	api "github.com/tvhung83/fshare/api"
	fshare "github.com/tvhung83/fshare/pkg"
	"gopkg.in/yaml.v2"
)

/*Config stores credentials and server config*/
type Config struct {
	Port     string `yaml:"port"`
	Username string `yaml:"user"`
	Password string `yaml:"pass"`
}

func processError(err error) {
	fmt.Println(err)
	os.Exit(2)
}

func readFile(cfg *Config) {
	f, err := os.Open("config.yml")
	if err != nil {
		processError(err)
	}

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
		processError(err)
	}
}

func main() {
	var cfg Config
	readFile(&cfg)
	api.Client = fshare.NewClient(cfg.Username, cfg.Password)
	http.HandleFunc("/file/", api.FileHandler)
	http.ListenAndServe(":"+cfg.Port, nil)
}
