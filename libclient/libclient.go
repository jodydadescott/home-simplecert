package libclient

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	hashauthrand "github.com/jodydadescott/simple-go-hash-auth/rand"
	hashauthserver "github.com/jodydadescott/simple-go-hash-auth/server"
)

type Client struct {
	secret     string
	url        string
	httpClient *http.Client
	token      *hashauthserver.Token
	rand       *hashauthrand.Rand
}

func New(config *Config) *Client {

	if config == nil {
		panic("config is nil")
	}

	if config.Secret == "" {
		panic("secret is empty")
	}

	if config.Server == "" {
		panic("url is empty")
	}

	return &Client{
		url:    config.Server,
		secret: config.Secret,
		rand:   hashauthrand.New(&hashauthrand.Config{}),
		httpClient: &http.Client{Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: config.SkipVerify},
		}},
	}
}

func (t *Client) Shutdown() {
	if t.httpClient != nil {
		t.httpClient.CloseIdleConnections()
	}
}

func (t *Client) GetCert(domain string) (*CR, error) {

	token, err := t.getToken()
	if err != nil {
		return nil, err
	}

	bearer := "Bearer " + token.Token

	params := url.Values{}
	params.Add("domain", domain)

	req, err := http.NewRequest(http.MethodGet, t.url+"/getcert?"+params.Encode(), nil)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Authorization", bearer)

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result CertResponse
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return nil, err
	}

	if result.Error != "" {
		return nil, fmt.Errorf(result.Error)
	}

	if result.CR == nil {
		return nil, fmt.Errorf("No CR in response")
	}

	return result.CR, nil
}

func (t *Client) getToken() (*Token, error) {

	if t.token != nil {
		if !isExpired(time.Now().Unix(), t.token.Exp) {
			return t.token, nil
		}
	}

	getAuthRequest := func() (*AuthRequest, error) {

		req, err := http.NewRequest(http.MethodGet, t.url+"/getauthrequest", nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := t.httpClient.Do(req)

		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
		}

		defer resp.Body.Close()

		bytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		var result AuthRequest
		err = json.Unmarshal(bytes, &result)
		if err != nil {
			return nil, err
		}

		result.ClientNonce = t.rand.String()
		hash := result.GetHashFromSecret(t.secret)
		result.Hash = hash
		return &result, nil
	}

	setToken := func() error {

		authRequest, err := getAuthRequest()
		if err != nil {
			return err
		}

		b, err := json.Marshal(authRequest)
		if err != nil {
			return err
		}

		req, err := http.NewRequest("GET", t.url+"/getauthtoken", bytes.NewBuffer(b))
		req.Header.Set("Content-Type", "application/json")

		if err != nil {
			return err
		}

		resp, err := t.httpClient.Do(req)
		if err != nil {
			return err
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("server returned status %d", resp.StatusCode)
		}

		defer resp.Body.Close()

		bytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		var result TokenResponse
		err = json.Unmarshal(bytes, &result)
		if err != nil {
			return err
		}

		if result.Error != "" {
			return fmt.Errorf(result.Error)
		}

		if result.Token == nil {
			return fmt.Errorf("No token in response")
		}

		if isExpired(time.Now().Unix(), result.Token.Exp) {
			return fmt.Errorf("Token already expired")
		}

		t.token = result.Token

		return nil
	}

	err := setToken()
	if err != nil {
		return nil, err
	}

	return t.token, nil
}

func isExpired(now, exp int64) bool {
	if now > exp {
		return true
	}
	return false
}
