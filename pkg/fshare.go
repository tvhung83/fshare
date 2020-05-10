package fshare

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"

	"github.com/avast/retry-go"
	"golang.org/x/net/publicsuffix"
)

/*Client wrapper type*/
type Client struct {
	Username string
	Password string
	Token    string
}

/*HTTPError captures status code and API's response body, in order to passthrough to client */
type HTTPError struct {
	StatusCode int
	Body       []byte
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("Status: %d, Body: %s", e.StatusCode, e.Body)
}

const baseURL = "https://api2.fshare.vn/api"
const filePrefix = "https://www.fshare.vn/file/"

var jar http.CookieJar
var client *http.Client

const loginPayload = `{
		"user_email": "%s",
		"password": "%s",
		"app_key": "GUxft6Beh3Bf8qKP7GC2IplYJZz1A53JQfRwne0R"
	}`

const downloadPayload = `{
		"token": "%s",
		"url": "%s"
	}`

const folderPayload = `{
		"token": "%s",
		"url": "%s",
		"dirOnly": 0,
		"pageIndex": %d,
		"limit": 1000
	}`

/*NewClient constructor & login*/
func NewClient(username string, password string) *Client {
	cl := &Client{Username: username, Password: password}
	cl.initHTTPClient()
	return cl
}

/*Login for first time (after construction) or re-login (after session expired)*/
func (c *Client) Login() error {
	return retry.Do(
		func() error {
			log.Printf("-- Logging %s", c.Username)

			body := []byte(fmt.Sprintf(loginPayload, c.Username, c.Password))
			body, statusCode, err := req("POST", "/user/login", body)
			err = c.wrapError(body, statusCode, err)
			if statusCode != 200 {
				return err
			}

			// grab login response and unmarshall to store token, session_id
			session := make(map[string]interface{})
			err = json.Unmarshal(body, &session)
			if err != nil {
				return &HTTPError{StatusCode: 500, Body: []byte(err.Error())}
			}
			c.Token = session["token"].(string)
			return nil
		},
		retry.Attempts(10),
		retry.Delay(10*time.Second),
		retry.OnRetry(func(n uint, err error) {
			log.Printf("Retry #%d in %d second(s), due to: %s", n+1, 10*(1<<n), err.Error())
		}),
	)
}

func perror(err error) {
	if err != nil {
		// panic(err)
		log.Fatal(err)
	}
}

func req(method string, path string, body []byte) ([]byte, int, error) {
	req, err := http.NewRequest(method, baseURL+path, bytes.NewBuffer(body))
	perror(err)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_2) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.4 Safari/605.1.15")
	req.Header.Set("Content-Type", "application/json")

	log.Printf("> %s to %s ...", method, path)
	resp, err := client.Do(req)
	perror(err)
	defer resp.Body.Close()

	log.Printf("< StatusCode: %d", resp.StatusCode)

	body, err = ioutil.ReadAll(resp.Body)
	return body, resp.StatusCode, err
}

/*IsLoggedIn checks if client is logged in*/
func (c *Client) IsLoggedIn() bool {
	body, statusCode, err := c.GetProfile()
	err = c.wrapError(body, statusCode, err)
	if statusCode != 200 || err != nil {
		log.Print(err)
		return false
	}
	profile := make(map[string]interface{})
	err = json.Unmarshal(body, &profile)
	return err == nil && profile != nil && profile["account_type"].(string) == "Vip"
}

/*GetProfile returns user profile if logged in*/
func (c *Client) GetProfile() ([]byte, int, error) {
	log.Printf("~~ Get user profile")
	return req("GET", "/user/get", []byte{})
}

/*Download get direct URL*/
func (c *Client) Download(url string) ([]byte, int, error) {
	// it's a file code
	if !strings.HasPrefix(url, filePrefix) {
		url = filePrefix + url
	}

	log.Printf("** Download %s", url)

	body := []byte(fmt.Sprintf(downloadPayload, c.Token, url))
	return req("POST", "/session/download", body)
}

/*GetFolder returns list of File*/
func (c *Client) GetFolder(url string, page int) ([]byte, int, error) {
	log.Printf("^^ Get Folder: %s", url)

	body := []byte(fmt.Sprintf(folderPayload, c.Token, url, page))
	return req("POST", "/fileops/getFolderList", body)
}

func (c *Client) initHTTPClient() {
	jar, _ = cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	client = &http.Client{
		Jar: jar,
	}
}

func (c *Client) wrapError(body []byte, statusCode int, err error) error {
	if statusCode != 200 {
		return &HTTPError{StatusCode: statusCode, Body: body}
	}
	if err != nil {
		return &HTTPError{StatusCode: 500, Body: []byte(err.Error())}
	}
	return nil
}
