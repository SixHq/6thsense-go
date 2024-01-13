package sixthGo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

/*query_args string,
timestamp int,
attack_type string,
status string,
learn_more_link string,
route string,*/

var config Config
var log_dict map[string]interface{}
var base_url string = "https://backend.withsix.co"
var apiKey string
var endpointss []string

func Initialize(endpoints []string, apikey string) {

	apiKey = apikey
	endpointss = endpoints
	url := base_url + "/project-config/config/" + apikey
	response, err := http.Get(url)
	// Check for errors
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer response.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	err = json.Unmarshal(body, &config)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	/*if config.RateLimiterEnabled {
		rateLimiteMiddleware(apikey, config, endpoints, log_dict, app)
	}*/
}

func ApplyMiddleWares(next http.Handler) http.Handler {
	if config.RateLimiterEnabled {
		return rateLimiteMiddleware(apiKey, config, endpointss, log_dict, next)
	} else {
		return next
	}
}

func SyncProjectRoutes(endpoints []string) (result map[string]interface{}, err error) {
	var rlConfigs = make(map[string]interface{})

	for _, route := range endpoints {
		editedRoute := strings.ReplaceAll(route, "/", "~")
		if config.BaseURL != "" && config.UserId != "" {
			if _, exists := config.RateLimiter[editedRoute]; exists {
				rlConfigs[editedRoute] = config.RateLimiter[editedRoute]
			} else {
				currentTime := time.Now()

				// Print the Unix timestamp in milliseconds
				unixMillis := currentTime.UnixNano() / int64(time.Millisecond)
				rlConfigs[editedRoute] = map[string]interface{}{
					"id":           editedRoute,
					"route":        editedRoute,
					"interval":     10,
					"rate_limit":   1,
					"last_updated": unixMillis,
					"created_at":   unixMillis,
					"unique_id":    "host",
					"is_active":    false,
				}
			}

		}
	}

	currentTime := time.Now()

	// Print the Unix timestamp in milliseconds
	unixMillis := currentTime.UnixNano() / int64(time.Millisecond)

	var newConfig = map[string]interface{}{
		"user_id":      config.UserId,
		"rate_limiter": rlConfigs,
		"encryption": map[string]interface{}{
			"public_key":   '_',
			"private_key":  '_',
			"use_count":    0,
			"last_updated": 0,
			"created_at":   unixMillis,
		},
		"base_url":             '_',
		"last_updated":         unixMillis,
		"created_at":           unixMillis,
		"encryption_enabled":   config.UserId != "" && config.EncryptionEnabled,
		"rate_limiter_enabled": config.UserId != "" && config.RateLimiterEnabled,
	}

	syncurl := base_url + "/project-config/config/sync-user-config"
	jsonBody, err := json.Marshal(newConfig)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}
	response, err := http.Post(syncurl, "application/json", bytes.NewBuffer(jsonBody))
	// Check for errors
	if err != nil {
		err = fmt.Errorf("Error:", err)
		return
	}
	defer response.Body.Close()

	result = newConfig
	return

}

func SyncProject(endpoints []string) {
	SyncProjectRoutes(endpoints)

	for _, route := range endpoints {
		editedRoute := strings.ReplaceAll(route, "/", "~")
		log_dict[editedRoute] = nil
	}
}
