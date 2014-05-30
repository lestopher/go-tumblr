package tumblr

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
)

const (
	// Version is the current version of this lib
	Version = "0.0.1"
	// BaseURL is the shared path to the tumblr api
	BaseURL = "http://api.tumblr.com/v2/"
	// UserAgent is the user agent when making requests
	UserAgent = "github.com/lestopher/tumblr v" + Version
)

// Client manages communication with Tumblr
type Client struct {
	// client  HTTP client to communicate with api
	client *http.Client

	// BaseURL URL used to communicate with api
	BaseURL *url.URL

	// UserAgent agent used with the api
	UserAgent string

	// ClientID the id of your oauth application
	ClientID string

	// ClientSecret the secret of your oauth application
	ClientSecret string

	// AccessToken token used when user is authenticated through oauth
	AccessToken string

	// Services used to communicate with the tumblr api
	Blog   *BlogService
	Users  *UsersService
	Tagged *TaggedService

	Response *Response
}

// ResponseMeta the meta response for requests
type ResponseMeta struct {
	Status int
	Msg    string
}

// Response the main response from tumblr api
type Response struct {
	HTTPResponse *http.Response
	Meta         *ResponseMeta `json:"meta, omitempty"`
	Response     interface{}   `json:"response, omitempty"`
}

// NewClient returns an initialized client
func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	baseURL, _ := url.Parse(BaseURL)

	c := &Client{
		client:    httpClient,
		BaseURL:   baseURL,
		UserAgent: UserAgent,
	}

	c.Users = &UsersService{client: c}
	c.Blog = &BlogService{client: c}
	c.Tagged = &TaggedService{client: c}

	return c
}

// Do sends an API request and returns the API response. The API response is
// decoded and stored in the value pointed to by v, or returned as an error if
// an API error has occurred.
func (c *Client) Do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	// err = CheckResponse(resp)
	// if err != nil {
	// 	return resp, err
	// }

	r := &Response{HTTPResponse: resp}
	if v != nil {
		r.Response = v
		err = json.NewDecoder(resp.Body).Decode(r)
		c.Response = r
	}
	return resp, err
}

// NewRequest creates an API request. A relative URL can be provided in urlStr,
// in which case it is resolved relative to the BaseURL of the Client.
// Relative URLs should always be specified without a preceding slash. If
// specified
func (c *Client) NewRequest(method, urlStr string, body string) (*http.Request, error) {
	rel, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	u := c.BaseURL.ResolveReference(rel)
	q := u.Query()
	if c.AccessToken != "" && q.Get("access_token") == "" {
		q.Set("access_token", c.AccessToken)
	}
	if c.ClientID != "" && q.Get("client_id") == "" {
		q.Set("client_id", c.ClientID)
	}
	if c.ClientSecret != "" && q.Get("client_secret") == "" {
		q.Set("client_secret", c.ClientSecret)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequest(method, u.String(), bytes.NewBufferString(body))
	if err != nil {
		return nil, err
	}

	if method == "POST" {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

	req.Header.Add("User-Agent", c.UserAgent)
	return req, nil
}

// CheckResponse checks the API response for error, and returns it
// if present. A response is considered an error if it has non StatusOK
// code.
// func CheckResponse(r *http.Response) error {
// 	if r.StatusCode == http.StatusOK {
// 		return nil
// 	}
//
// 	resp := new(ErrorResponse)
// 	resp.Response = r
//
// 	// Sometimes Instagram returns 500 with plain message
// 	// "Oops, an error occurred.".
// 	if r.StatusCode == http.StatusInternalServerError {
// 		meta := &ResponseMeta{
// 			ErrorType:    "Internal Server Error",
// 			Code:         500,
// 			ErrorMessage: "Oops, an error occurred.",
// 		}
// 		resp.Meta = meta
//
// 		return resp
// 	}
//
// 	data, err := ioutil.ReadAll(r.Body)
// 	if err == nil && data != nil {
// 		json.Unmarshal(data, resp)
// 	}
// 	return resp
// }
