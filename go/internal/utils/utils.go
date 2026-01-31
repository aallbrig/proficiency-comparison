package utils

import (
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
		return string(start)
	}
	return string(start) + "-" + string(end)
}
