package seshandler

import (
	"encoding/json"
	"errors"
	log "gitlab.starlink.ua/high-school-prod/chat/logger"
	"io/ioutil"
	"net/http"
	"os"
)

// APIResponse - a response from the API with all of the user information.
type APIResponse struct {
	Ok        bool   `json:"ok,bool,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
	Data      struct {
		User struct {
			ID       int    `json:"id,omitempty"`
			Username string `json:"username,omitempty"`
			Email    string `json:"email,omitempty"`
		}
	}
	Error struct {
		Message string `json:"message,omitempty"`
	}
}

// sendRequest - will send a prepared HTTP request with the user token.
func sendRequest(url, method, token string) *http.Response {
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err == nil {
		req.Header.Add("Authorization", token)
		resp, err := client.Do(req)
		if err == nil {
			return resp
		} else {
			log.Logger.Error(err)
			return nil
		}
	} else {
		log.Logger.Error(err)
		return nil
	}
}

// fetchUser - will fetch user info from API with the given user token.
func fetchUser(token string) (*APIResponse, error) {
	var apiResponse *APIResponse = nil
	apiRoot, _ := os.LookupEnv("API_BASE_URL")
	resp := sendRequest(apiRoot+"/user", "POST", token)

	// Check that the response wasn't empty.
	if resp == nil {
		log.Logger.Error("API request yielded empty response")
		return nil, errors.New("API request yielded empty response")
	}

	// Check that http status is ok.
	if resp.StatusCode != http.StatusOK {
		log.Logger.Error("API request yielded HTTP not ok")
		return nil, errors.New("API request yielded HTTP not ok")
	}

	// Handle ioutil error.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Logger.Error(err)
		return nil, err
	}

	// Try to unmarshal.
	apiResponse = &APIResponse{}
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		log.Logger.Error(err)
		apiResponse = nil
	}

	// Checking 'ok' field in response body's JSON.
	if !apiResponse.Ok {
		err = errors.New(apiResponse.Error.Message)
		log.Logger.Warn(err)
		return nil, err
	}

	return apiResponse, nil
}
