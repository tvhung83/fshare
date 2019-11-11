package fshare

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

/*LoginResponse contains info about session*/
type LoginResponse struct {
	Code      int    `json:"code"`
	Msg       string `json:"msg"`
	Token     int    `json:"token"`
	SessionID int    `json:"session_id"`
}

/*File is an element of folder listing response*/
type File struct {
	ID            int       `json:"id"`
	LinkCode      string    `json:"linkcode"`
	Name          string    `json:"name"`
	Type          int       `json:"type"`
	Path          string    `json:"path"`
	Size          int64     `json:"size"`
	DownloadCount int       `json:"downloadcount"`
	Deleted       bool      `json:"deleted"`
	MimeType      string    `json:"mimetype"`
	Created       time.Time `json:"created"`
	Modified      time.Time `json:"modified"`
	Modified2     time.Time `json:"modified2"`
	Realname      string    `json:"realname"`
}

var client = &http.Client{}
var sess *LoginResponse

func perror(err error) {
	if err != nil {
		panic(err)
	}
}

func req(method string, url string, body *bytes.Buffer) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	perror(err)
	req.Header.Set("User-Agent", "okhttp/3.6.0")
	req.Header.Set("Content-Type", "application/json")

	log.Printf("> %s to %s ...", method, url)
	return client.Do(req)
}

/*Login to obtain token and session_id*/
func Login(username string, password string) (*LoginResponse, error) {
	payload := `{
		"user_email": "%s",
		"password": "%s",
		"app_key": "L2S7R6ZMagggC5wWkQhX2+aDi467PPuftWUMRFSn"
	}`

	log.Printf("-- Logging %s", username)
	resp, err := req("POST", "https://api.fshare.vn/api/user/login/", bytes.NewBufferString(fmt.Sprintf(payload, username, password)))
	perror(err)
	defer resp.Body.Close()
	log.Printf("< StatusCode: %d", resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	perror(err)

	var r = new(LoginResponse)
	err = json.Unmarshal(body, &r)
	return r, err
}

/*Download get direct URL*/
func Download(url string) (string, error) {
	prefix := "https://www.fshare.vn/file/"
	// it's a file code
	if !strings.HasPrefix(url, prefix) {
		url = prefix + url
	}
	payload := `{
		"token": "%s",
		"url": %s
	}`

	log.Printf("** Download %s", url)
	resp, err := req("POST", "https://api.fshare.vn/api/session/download", bytes.NewBufferString(fmt.Sprintf(payload)))
	perror(err)
	defer resp.Body.Close()
	log.Printf("< StatusCode: %d", resp.StatusCode)

	if resp.StatusCode == 403 {
		return "", errors.New("Permission denied")
	}

	if resp.StatusCode != 200 {
		return "", errors.New("Dead link")
	}

	body, err := ioutil.ReadAll(resp.Body)
	perror(err)

	result := make(map[string]interface{})
	err = json.Unmarshal(body, &result)
	return string(result["location"].([]byte)), err
}

/*GetFolder returns list of File*/
func GetFolder(url string, page int) (*[]File, error) {
	payload := `{
		"token": "%s",
		"url": "%s",
		"dirOnly": 0,
		"pageIndex": 0,
		"limit": 1000
	}`

	log.Printf("^^ Get Folder: %s", url)
	resp, err := req("POST", "https://api.fshare.vn/api/fileops/getFolderList", bytes.NewBufferString(fmt.Sprintf(payload, sess.Token, url)))
	perror(err)
	defer resp.Body.Close()
	log.Printf("< StatusCode: %d", resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	perror(err)

	var r = new([]File)
	err = json.Unmarshal(body, &r)
	return r, err
}
