package validator

import (
	"fmt"
	"net/url"
	"strings"
)

const MaxURLLength = 2048

func ValidateURL(rawURL string) error {
	if rawURL == "" {
		return fmt.Errorf("url is required")
	}

	if len(rawURL) > MaxURLLength {
		return fmt.Errorf("url exceeds maximum length of %d characters", MaxURLLength)
	}

	parsedURL, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL")
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("only http and https URLs are supported")
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("URL must contain a host")
	}

	if strings.TrimSpace(parsedURL.Host) == "" {
		return fmt.Errorf("URL must contain a valid host")
	}

	return nil
}
