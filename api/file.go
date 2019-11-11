package api

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

/*FileHandler returns direct URL from raw URL*/
func FileHandler(w http.ResponseWriter, r *http.Request) {
	resp, _ := http.Get("https://www.fshare.vn/site/login")
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprintf(w, string(b))
}
