package sixthGo

import (
	"fmt"
	"time"
)

func getTimeNow() float64 {
	location, err := time.LoadLocation("Africa/Lagos")
	if err != nil {
		// Handle error
		fmt.Println("Error loading location:", err)
		return 0.0
	}

	now := time.Now().In(location)
	return float64(now.Unix())
}
