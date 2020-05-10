package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	fshare "github.com/tvhung83/fshare/pkg"
)

/*Client is fshare Client type, we put it here to share with main*/
var Client *fshare.Client

/*FileHandler returns direct URL from raw URL*/
func FileHandler(w http.ResponseWriter, r *http.Request) {
	var id string
	if r.Method == http.MethodGet {
		id = strings.TrimPrefix(r.URL.Path, "/file/")
	} else if r.Method == http.MethodPost {
		r.ParseForm()
		id = r.FormValue("file")
	}

	body, statusCode, err := Client.Download(id)
	if err == nil && statusCode == 200 {
		result := make(map[string]string)
		err = json.Unmarshal(body, &result)
		url := fmt.Sprint(result["location"])
		log.Printf("[OK] %s >> %s", id, url)

		if r.Method == http.MethodGet {
			http.Redirect(w, r, url, http.StatusFound)
		} else if r.Method == http.MethodPost {
			w.WriteHeader(200)
			w.Write([]byte(url))
		}
	} else {
		writeTo(w, body, statusCode, err)
	}
}

/*FolderHandler returns direct URL from raw URL*/
func FolderHandler(w http.ResponseWriter, r *http.Request) {
	var id string
	if r.Method == http.MethodGet {
		id = strings.TrimPrefix(r.URL.Path, "/folder/")
	} else if r.Method == http.MethodPost {
		r.ParseForm()
		id = r.FormValue("folder")
	}
	body, statusCode, err := Client.GetFolder(id, 0)
	writeTo(w, body, statusCode, err)
}

/*Ping check login status */
func Ping(w http.ResponseWriter, r *http.Request) {
	if Client.IsLoggedIn() {
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	} else {
		w.WriteHeader(500)
		w.Write([]byte("NOK"))
	}
}

/*Login force re-login */
func Login(w http.ResponseWriter, r *http.Request) {
	err := Client.Login()
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
	} else {
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	}
}

func writeTo(w http.ResponseWriter, body []byte, statusCode int, err error) {
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		w.Write(body)
	}
}
