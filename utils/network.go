package utils

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// HTTPGetAsBytes performs an HTTP GET request to the specified URL and returns the response body as a byte slice.
//
// @param url The URL to send the HTTP GET request to.
// @param retry The number of times to retry the request if it fails or encounters a rate limit.
// @param initialWaitingSeconds The initial number of seconds to wait before retrying the request.
//
// @return The response body as a byte slice.
func HTTPGetAsBytes(url string, retry int8, initialWaitingSeconds int32) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		if retry > 0 {
			time.Sleep(time.Duration(initialWaitingSeconds) * time.Second)
			return HTTPGetAsBytes(url, retry-1, initialWaitingSeconds*2)
		}
		return []byte{}, fmt.Errorf("failed to GET the URL '%s': %w", url, err)
	}
	if resp.StatusCode > 399 {
		if resp.Header.Get("x-ratelimit-remaining") == "0" {
			time.Sleep(time.Until(time.Unix(ParseInt(resp.Header.Get("x-ratelimit-reset"))+10, 0)))
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
// @param initial_waiting_seconds The initial number of seconds to wait before retrying the request.
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
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to GET the url '%s': %w", url, err)
	}
	defer resp.Body.Close()
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to write the response body to '%s': %w", name, err)
	}
	return out.Name(), nil
}
