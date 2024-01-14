package sixthGo

type Encryption struct {
	CreatedAt   int    `json:"created_at"`
	LastUpdated int    `json:"last_updated"`
	PrivateKey  string `json:"private_key"`
	PublicKey   string `json:"public_key"`
	UseCount    int    `json:"use_count"`
}

type ErrorPayload map[string]interface{}

type RateLimiter struct {
	ID             string                  `json:"id"`
	Route          string                  `json:"route"`
	Interval       int                     `json:"interval"`
	RateLimit      int                     `json:"rate_limit"`
	LastUpdated    float64                 `json:"last_updated"`
	IsActive       bool                    `json:"is_active"`
	UniqueID       string                  `json:"unique_id"`
	ErrorPayloadID string                  `json:"error_payload_id"`
	ErrorPayload   map[string]ErrorPayload `json:"error_payload"`
	RateLimitType  string                  `json:"rate_limit_type"`
}

type Config struct {
	BaseURL            string                 `json:"base_url"`
	CreatedAt          float64                `json:"created_at"`
	Encryption         Encryption             `json:"encryption"`
	EncryptionEnabled  bool                   `json:"encryption_enabled"`
	LastUpdated        float64                `json:"last_updated"`
	RateLimiter        map[string]RateLimiter `json:"rate_limiter"`
	RateLimiterEnabled bool                   `json:"rate_limiter_enabled"`
	UserId             string                 `json:"user_id"`
}
