package sixthGo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type MyResponse struct {
	Response bool `json:"response"`
	// Add other fields based on your actual JSON structure
}

func rateLimiteMiddleware(apikey string, config Config, endpoints []string, log_dict map[string]interface{}, next http.Handler) http.Handler {

	isRateLimitReached := func(config Config, uid, route string) (bool, error) {

		rateLimit := config.RateLimiter[route].RateLimit
		interval := config.RateLimiter[route].Interval

		body := map[string]interface{}{
			"route":      route,
			"interval":   interval,
			"rate_limit": rateLimit,
			"unique_id":  strings.ReplaceAll(uid, ".", "~"),
			"user_id":    config.UserId,
			"is_active":  true,
		}

		jsonBody, _ := json.Marshal(body)

		// Assuming you have a function similar to axios.post for making HTTP POST requests
		response, err := http.Post("https://backend.withsix.co/rate-limit/enquire-has-reached-rate_limit", "application/json", bytes.NewBuffer(jsonBody))
		if err != nil {
			return false, err
		}
		defer response.Body.Close()

		if response.StatusCode == 200 {
			bod, err := ioutil.ReadAll(response.Body)
			if err != nil {
				return false, err
			}

			var jsonResponse MyResponse
			err = json.Unmarshal(bod, &jsonResponse)
			if err != nil {
				fmt.Println("Error unmarshaling JSON:", err)
				return false, err
			}

			return jsonResponse.Response, nil
		}
		return false, nil
	}

	sendlogs := func(apiKey string, logDict map[string]interface{}, route string, header map[string][]string, body interface{}, query map[string][]string) error {
		timestamp := getTimeNow()
		lastLogSentRaw, exists := logDict[route]

		// Check if the key exists in the map
		if !exists {
			lastLogSentRaw = nil
		}

		// Convert lastLogSent to int64
		var lastLogSent float64
		if lastLogSentRaw != nil {
			lastLogSent, _ = lastLogSentRaw.(float64)
		}

		if lastLogSent == 0 || timestamp-lastLogSent > 10000 {
			payload := map[string]interface{}{
				"header":          header,
				"user_id":         apiKey,
				"body":            body,
				"query_args":      query,
				"timestamp":       timestamp,
				"attack_type":     "No Rate Limit Attack",
				"cwe_link":        "https://cwe.mitre.org/data/definitions/770.html",
				"status":          "MITIGATED",
				"learn_more_link": "https://en.wikipedia.org/wiki/Rate_limiting",
				"route":           route,
			}

			// Convert payload to JSON
			jsonBody, err := json.Marshal(payload)
			if err != nil {
				return err
			}

			// Make the HTTP POST request
			response, err := http.Post("https://backend.withsix.co/slack/send_message_to_slack_user", "application/json", bytes.NewBuffer(jsonBody))
			if err != nil {
				return err
			}
			defer response.Body.Close()

			// Handle the response as needed
			// ...

			logDict[route] = timestamp
		}

		return nil
	}

	// Utility function to get the preferred ID based on rate limit type
	getPreferredID := func(req *http.Request, rateLimitResponse RateLimiter) string {
		preferredID := rateLimitResponse.UniqueID

		if preferredID == "" || preferredID == "host" {
			return req.Header.Get("Host")
		}

		if rateLimitResponse.RateLimitType == "body" {
			preferredID = req.FormValue(preferredID)
		} else if rateLimitResponse.RateLimitType == "header" {
			preferredID = req.Header.Get(preferredID)
		} else if rateLimitResponse.RateLimitType == "args" {
			preferredID = req.URL.Query().Get(preferredID)
		}

		return preferredID
	}

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		fmt.Println("kilode3")
		// Middleware logic
		//host := req.Header.Get("Host")
		statusCode := 200
		route := strings.ReplaceAll(req.URL.Path, "/", "~")
		var body interface{} // Change the type based on your actual payload structure

		// Fail-safe in case the sixth server is down
		tryMiddlewareLogic := func() {
			updatedTime := getTimeNow()

			if updatedTime-config.RateLimiter[route].LastUpdated > 60000 {
				response, err := http.Get("https://backend.withsix.co/project-config/config/get-route-rate-limit/" + apikey + "/" + route)
				fmt.Println(response.StatusCode, "kilode4", http.StatusOK)
				if err == nil && response.StatusCode == http.StatusOK {
					statusCode = response.StatusCode
					if lastTimeUpdated, ok := config.RateLimiter[route]; ok {
						lastTimeUpdated.LastUpdated = updatedTime
						config.RateLimiter[route] = lastTimeUpdated
					} else {
						next.ServeHTTP(w, req)
					}

					if statusCode == http.StatusOK {
						fmt.Println("kilode2")
						if response.Header.Get("Content-Type") == "application/json" {
							decoder := json.NewDecoder(response.Body)
							var rateLimitResponse RateLimiter
							if err := decoder.Decode(&rateLimitResponse); err == nil {

								config.RateLimiter[route] = rateLimitResponse
								preferredID := getPreferredID(req, rateLimitResponse)
								result, err := isRateLimitReached(config, preferredID, route)
								if err != nil {
									fmt.Println("Error marshaling JSON:", err)
									next.ServeHTTP(w, req)
								}

								if result {
									fmt.Println("kilode1")
									sendlogs(apikey, log_dict, route, req.Header, body, req.URL.Query())
									tempPayload := rateLimitResponse.ErrorPayload
									final := make(map[string]interface{})

									for _, value := range tempPayload {
										for key, val := range value {
											if key != "uid" {
												final[key] = val
											}
										}
									}

									stringed, _ := json.Marshal(final)
									newHeader := map[string]string{
										"Content-Length": fmt.Sprintf("%d", len(stringed)),
										"Content-Type":   "application/json",
									}
									for key, value := range newHeader {
										w.Header().Set(key, value)
									}
									w.WriteHeader(http.StatusBadRequest)
									w.Write(stringed)

								} else {
									next.ServeHTTP(w, req)
								}
							} else {
								next.ServeHTTP(w, req)
							}
						} else {
							next.ServeHTTP(w, req)
						}
					} else {
						next.ServeHTTP(w, req)
					}
				}
			}
			next.ServeHTTP(w, req)
		}

		tryMiddlewareLogic()
	})
}
