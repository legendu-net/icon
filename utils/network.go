package utils

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func IsErrorHTTPResponse(resp *http.Response) bool {
	//nolint:mnd // good readability in the context
	return resp.StatusCode >= 400
}

// HTTPGetAsBytes performs an HTTP GET request to the specified URL and returns the response body as a byte slice.
//
// @param url The URL to send the HTTP GET request to.
// @param retry The number of times to retry the request if it fails or encounters a rate limit.
// @param initialWaitingSeconds The initial number of seconds to wait before retrying the request.
//
// @return The response body as a byte slice.
func HTTPGetAsBytes(url string, retry int8, initialWaitingSeconds int32) ([]byte, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, http.NoBody)
	if err != nil {
		if retry > 0 {
			time.Sleep(time.Duration(initialWaitingSeconds) * time.Second)
			return HTTPGetAsBytes(url, retry-1, initialWaitingSeconds*2)
		}
		return []byte{}, fmt.Errorf("failed to create a HTTP GET request to the URL '%s' with context: %w", url, err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if retry > 0 {
			time.Sleep(time.Duration(initialWaitingSeconds) * time.Second)
			return HTTPGetAsBytes(url, retry-1, initialWaitingSeconds*2)
		}
		return []byte{}, fmt.Errorf("the HTTP GET request to the URL '%s' failed: %w", url, err)
	}
	if IsErrorHTTPResponse(resp) {
		if resp.Header.Get("x-ratelimit-remaining") == "0" {
			rateLimitResetInSeconds := ParseInt(resp.Header.Get("x-ratelimit-reset"))
			//nolint:mnd // good readability in the context
			rateLimitResetTime := time.Unix(rateLimitResetInSeconds+10, 0)
			time.Sleep(time.Until(rateLimitResetTime))
			return HTTPGetAsBytes(url, retry, initialWaitingSeconds)
		}
		if retry > 0 {
			time.Sleep(time.Duration(initialWaitingSeconds) * time.Second)
			return HTTPGetAsBytes(url, retry-1, initialWaitingSeconds*2)
		}
		return []byte{}, fmt.Errorf(
			`the HTTP GET request on the URL %s got an error response with the status code %d.
			x-ratelimit-limit: %s 
			x-ratelimit-remaining: %s 
			x-ratelimit-reset: %s`,
			url,
			resp.StatusCode,
			resp.Header.Get("x-ratelimit-limit"),
			resp.Header.Get("x-ratelimit-remaining"),
			time.Unix(ParseInt(resp.Header.Get("x-ratelimit-reset")), 0).Local().Format(time.RFC3339),
		)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		if retry > 0 {
			time.Sleep(time.Duration(initialWaitingSeconds) * time.Second)
			return HTTPGetAsBytes(url, retry-1, initialWaitingSeconds*2)
		}
		return []byte{}, fmt.Errorf("failed to read the response body: %w", err)
	}
	return body, nil
}

// HttpGetAsString performs an HTTP GET request to the specified URL and returns the response body as a string.
//
// @param url The URL to send the HTTP GET request to.
// @param retry The number of times to retry the request if it fails or encounters a rate limit.
// @param initialWaitingSeconds The initial number of seconds to wait before retrying the request.
//
// @return The response body as a string.
func HTTPGetAsString(url string, retry int8, initialWaitingSeconds int32) string {
	bytes, err := HTTPGetAsBytes(url, retry, initialWaitingSeconds)
	if err != nil {
		log.Fatal(err)
	}
	return string(bytes)
}

// Download file from the given URL.
//
// @param url The URL of the file to download.
// @param name The desired name of the file when saved locally.
// @param useTempDir If true, the file will be saved to a temporary directory.
//
// @return The local path where the downloaded file is saved.
func DownloadFile(url, name string, useTempDir bool) (string, error) {
	var out *os.File
	var err error
	if useTempDir {
		name = filepath.Join(CreateTempDir(""), name)
	}
	out, err = os.Create(name)
	if err != nil {
		return "", fmt.Errorf("failed to create file '%s': %w", name, err)
	}
	defer out.Close()
	log.Printf("Downloading %s to %s\n", url, out.Name())
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, http.NoBody)
	if err != nil {
		return "", fmt.Errorf("failed to create a HTTP GET request to the URL '%s' with context: %w", url, err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("the HTTP GET request to the URL '%s' failed: %w", url, err)
	}
	defer resp.Body.Close()
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to write the response body to '%s': %w", name, err)
	}
	return out.Name(), nil
}
