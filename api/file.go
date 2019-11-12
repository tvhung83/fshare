package api

import (
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
	directURL, err := Client.Download(id)
	if err != nil {
		log.Fatal(err)
	}

	url := fmt.Sprint(directURL)
	log.Printf("[OK] %s >> %s", id, url)

	if r.Method == http.MethodGet {
		http.Redirect(w, r, url, http.StatusFound)
	} else if r.Method == http.MethodPost {
		w.WriteHeader(200)
		w.Write([]byte(url))
	}
}
