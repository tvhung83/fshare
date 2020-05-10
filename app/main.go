package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/tvhung83/fshare/api"
	fshare "github.com/tvhung83/fshare/pkg"
	"gopkg.in/robfig/cron.v2"
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
	ex, err := os.Executable()
	if err != nil {
		processError(err)
	}
	exPath := filepath.Dir(ex)
	f, err := os.Open(exPath + "/config.yml")
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
	api.Client.Login()

	// init cron job to check for logged in status and re-login
	c := cron.New()
	c.AddFunc("@hourly", func() {
		if !api.Client.IsLoggedIn() {
			api.Client.Login()
		}
	})
	c.Start()

	http.HandleFunc("/ping", api.Ping)
	http.HandleFunc("/login", api.Login)
	http.HandleFunc("/file/", api.FileHandler)
	http.HandleFunc("/folder/", api.FolderHandler)
	http.ListenAndServe(":"+cfg.Port, nil)
}
