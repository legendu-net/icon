package utils

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// HttpGetAsBytes performs an HTTP GET request to the specified URL and returns the response body as a byte slice.
//
// @param url The URL to send the HTTP GET request to.
// @param retry The number of times to retry the request if it fails or encounters a rate limit.
// @param initial_waiting_seconds The initial number of seconds to wait before retrying the request.
//
// @return The response body as a byte slice.
func HttpGetAsBytes(url string, retry int8, initial_waiting_seconds int32) []byte {
	resp, err := http.Get(url)
	if err != nil {
		if retry > 0 {
			time.Sleep(time.Duration(initial_waiting_seconds) * time.Second)
			return HttpGetAsBytes(url, retry-1, initial_waiting_seconds*2)
		}
		log.Fatal("The HTTP GET request on the URL ", url, " got the following error:\n", err)
	}
	if resp.StatusCode > 399 {
		if resp.Header.Get("x-ratelimit-remaining") == "0" {
			time.Sleep(time.Until(time.Unix(ParseInt(resp.Header.Get("x-ratelimit-reset"))+10, 0)))
			return HttpGetAsBytes(url, retry, initial_waiting_seconds)
		}
		if retry > 0 {
			time.Sleep(time.Duration(initial_waiting_seconds) * time.Second)
			return HttpGetAsBytes(url, retry-1, initial_waiting_seconds*2)
		}
		log.Fatal(
			"The HTTP GET request on the URL ", url, " got an error response with the status code ",
			resp.StatusCode,
			"\n",
			"x-ratelimit-limit: ",
			resp.Header.Get("x-ratelimit-limit"),
			"\n",
			"x-ratelimit-remaining: ",
			resp.Header.Get("x-ratelimit-remaining"),
			"\n",
			"x-ratelimit-reset: ",
			time.Unix(ParseInt(resp.Header.Get("x-ratelimit-reset")), 0).Local(),
			"\n",
		)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		if retry > 0 {
			time.Sleep(time.Duration(initial_waiting_seconds) * time.Second)
			return HttpGetAsBytes(url, retry-1, initial_waiting_seconds*2)
		}
		log.Fatal("Reading the response body of the http GET request on the url ", url, " got the following error:\n", err)
	}
	return body
}

// HttpGetAsString performs an HTTP GET request to the specified URL and returns the response body as a string.
//
// @param url The URL to send the HTTP GET request to.
// @param retry The number of times to retry the request if it fails or encounters a rate limit.
// @param initial_waiting_seconds The initial number of seconds to wait before retrying the request.
//
// @return The response body as a string.
func HttpGetAsString(url string, retry int8, initial_waiting_seconds int32) string {
	return string(HttpGetAsBytes(url, retry, initial_waiting_seconds))
}

// Download file from the given URL.
//
// @param url The URL of the file to download.
// @param name The desired name of the file when saved locally.
// @param useTempDir If true, the file will be saved to a temporary directory.
//
// @return The local path where the downloaded file is saved.
func DownloadFile(url string, name string, useTempDir bool) string {
	var out *os.File
	var err error
	if useTempDir {
		name = filepath.Join(CreateTempDir(""), name)
	}
	out, err = os.Create(name)
	if err != nil {
		log.Fatal("ERROR - ", err, ": ", name)
	}
	defer out.Close()
	log.Printf("Downloading %s to %s\n", url, out.Name())
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	defer resp.Body.Close()
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Fatal("ERROR - ", err)
	}
	return out.Name()
}
