package apiextensions

import (
	"fmt"
	"net/url"
)

func validationError(targetURL string) error {
	return fmt.Errorf("invalid URL: %s, must be an absolute URL", targetURL)
}

func ValidateAbsoluteURL(targetURL string) (*url.URL, error) {
	parsedURL, err := url.ParseRequestURI(targetURL)
	if err != nil {
		return nil, validationError(targetURL)
	}

	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return nil, validationError(targetURL)
	}

	return parsedURL, nil
}
