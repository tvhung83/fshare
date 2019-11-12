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

/*LoginResponse /api/user/login*/
type LoginResponse struct {
	Code      int    `json:"code"`
	Msg       string `json:"msg"`
	Token     string `json:"token"`
	SessionID string `json:"session_id"`
}

/*FileData /api/fileops/getFolderList*/
type FileData struct {
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

/*Profile /api/user/get */
type Profile struct {
	ID                    int    `json:"id"`
	Level                 int    `json:"level"`
	Name                  string `json:"name"`
	Phone                 string `json:"phone"`
	Birthday              string `json:"birthday"`
	Gender                string `json:"gender"`
	Address               string `json:"address"`
	IDCard                string `json:"id_card"`
	City                  string `json:"city"`
	Occupation            string `json:"occupation"`
	Email                 string `json:"email"`
	JoinDate              uint32 `json:"joindate"`
	TotalPoints           uint   `json:"totalpoints"`
	ExpireVip             uint32 `json:"expire_vip"`
	Traffic               uint32 `json:"traffic"`
	TrafficUsed           uint32 `json:"traffic_used"`
	Webspace              uint32 `json:"webspace"`
	WebspaceUsed          uint32 `json:"webspace_used"`
	WebspaceSecure        uint32 `json:"webspace_secure"`
	WebspaceSecureUsed    uint32 `json:"webspace_secure_used"`
	Amount                uint32 `json:"amount"`
	DLTimeAvailable       uint32 `json:"dl_time_avail"`
	StatusTelesalePrepaid int    `json:"status_telesale_prepaid"`
	AccountType           string `json:"account_type"`
}

/*Client wrapper type*/
type Client struct {
	Username string
	Password string
	Session  *LoginResponse
}

func (c *Client) login() {
	var err error
	c.Session, err = Login(c.Username, c.Password)
	perror(err)
}

/*NewClient constructor & login*/
func NewClient(username string, password string) *Client {
	cl := &Client{Username: username, Password: password}
	cl.login()
	return cl
}

const baseURL = "https://api.fshare.vn/api"
const filePrefix = "https://www.fshare.vn/file/"

var client = &http.Client{}

func perror(err error) {
	if err != nil {
		panic(err)
	}
}

func req(method string, path string, body []byte, sessionID string) ([]byte, error) {
	req, err := http.NewRequest(method, baseURL+url, bytes.NewBuffer(body))
	perror(err)
	req.Header.Set("User-Agent", "okhttp/3.6.0")
	req.Header.Set("Content-Type", "application/json")
	if sessionID != "" {
		req.Header.Set("cookie", fmt.Sprintf("session_id=%s", sessionID))
	}

	log.Printf("> %s to %s ...", method, url)
	resp, err := client.Do(req)
	perror(err)
	defer resp.Body.Close()

	log.Printf("< StatusCode: %d", resp.StatusCode)

	if resp.StatusCode == 403 {
		return nil, errors.New("Permission denied")
	}

	return ioutil.ReadAll(resp.Body)
}

const loginPayload = `{
		"user_email": "%s",
		"password": "%s",
		"app_key": "L2S7R6ZMagggC5wWkQhX2+aDi467PPuftWUMRFSn"
	}`

/*Login to obtain token and session_id*/
func Login(username string, password string) (*LoginResponse, error) {
	log.Printf("-- Logging %s", username)

	body := []byte(fmt.Sprintf(loginPayload, username, password))
	body, err := req("POST", "/user/login", body, "")
	perror(err)

	var session = new(LoginResponse)
	err = json.Unmarshal(body, &session)
	return session, err
}

/*IsLoggedIn checks if client is logged in*/
func (c *Client) IsLoggedIn() bool {
	return c.Session != nil && c.Session.Code == 200
}

/*IsVip checks if account type is 'Vip'*/
func (c *Client) IsVip() bool {
	if !c.IsLoggedIn() {
		return false
	}
	profile, err := c.GetProfile()
	perror(err)
	return profile != nil && profile.AccountType == "Vip"
}

/*GetProfile returns user profile if logged in*/
func (c *Client) GetProfile() (*Profile, error) {
	if !c.IsLoggedIn() {
		return nil, errors.New("Not logged in")
	}

	body, err := req("GET", "/user/get", []byte{}, c.Session.SessionID)
	perror(err)

	var profile = new(Profile)
	err = json.Unmarshal(body, &profile)
	return profile, err
}

const downloadPayload = `{
		"token": "%s",
		"url": "%s"
	}`

/*Download get direct URL*/
func (c *Client) Download(url string) (interface{}, error) {
	// it's a file code
	if !strings.HasPrefix(url, filePrefix) {
		url = filePrefix + url
	}

	log.Printf("** Download %s", url)

	body := []byte(fmt.Sprintf(downloadPayload, c.Session.Token, url))
	body, err := req("POST", "/session/download", body, c.Session.SessionID)
	perror(err)

	result := make(map[string]interface{})
	err = json.Unmarshal(body, &result)
	return result["location"], err
}

const folderPayload = `{
		"token": "%s",
		"url": "%s",
		"dirOnly": 0,
		"pageIndex": %d,
		"limit": 1000
	}`

/*GetFolder returns list of File*/
func (c *Client) GetFolder(url string, page int) (*[]FileData, error) {

	log.Printf("^^ Get Folder: %s", url)

	body := []byte(fmt.Sprintf(folderPayload, c.Session.Token, url, page))
	body, err := req("POST", "/fileops/getFolderList", body, c.Session.SessionID)
	perror(err)

	var files = new([]FileData)
	err = json.Unmarshal(body, &files)
	return files, err
}
