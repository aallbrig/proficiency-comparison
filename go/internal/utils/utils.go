package utils

import (
	"fmt"
	"net/http"
	"time"
)

func CheckConnectivity(url string) bool {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	
	resp, err := client.Head(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	
	return resp.StatusCode < 400
}

func FormatYearRange(start, end int) string {
	if start == end {
		return fmt.Sprintf("%d", start)
	}
	return fmt.Sprintf("%d-%d", start, end)
}
