package main

import (
	"io"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

func main() {
	// cors proxy
	skippedHeaders := []string{
		"Host",
		"Origin",
		"Referer",
		"Access-Control-Request-Method",
		"Access-Control-Request-Headers",
		"X-Cookie",
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		cors(w, r)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		u := r.URL.Query().Get("url")
		if u == "" {
			http.Error(w, "missing 'url' query parameter", http.StatusBadRequest)
			return
		}
		parsedURL, err := url.Parse(u)
		if err != nil {
			http.Error(w, "invalid 'url' query parameter: "+err.Error(), http.StatusBadRequest)
			return
		}
		if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
			http.Error(w, "invalid 'url' query parameter: only http and https are allowed", http.StatusBadRequest)
			return
		}

		// tie the request to the context of the original request to handle cancellation
		req, err := http.NewRequestWithContext(r.Context(), r.Method, u, r.Body)
		if err != nil {
			http.Error(w, "failed to create request: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// add the X-Cookie header if present in the original request
		if cookie := r.Header.Get("X-Cookie"); cookie != "" {
			req.Header.Set("Cookie", cookie)
		}

		// copy headers from the original request, skipping specified headers
		for k, v := range r.Header {
			if !slices.Contains(skippedHeaders, k) {
				req.Header[k] = v
			}
		}

		// x-forwarded-for header
		if clientIP := r.RemoteAddr; clientIP != "" {
			req.Header.Set("X-Forwarded-For", clientIP)
		}

		// timeout for the request
		var timeout int64 = 30 // default timeout in seconds
		timeoutRaw := r.URL.Query().Get("timeout")
		if timeoutRaw != "" {
			timeout, err = strconv.ParseInt(timeoutRaw, 10, 64)
			if err != nil {
				http.Error(w, "invalid 'timeout' query parameter: "+err.Error(), http.StatusBadRequest)
				return
			}
		}

		client := &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		}

		resp, err := client.Do(req)
		for name, values := range resp.Header {
			if strings.HasPrefix(strings.ToLower(name), "access-control-") { // strip upstream CORS headers to prevent duplicates
				continue
			}
			for _, value := range values {
				w.Header().Add(name, value)
			}
		}
		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(w, resp.Body)
	})
	port := os.Getenv("PORT")
	if port == "" {
		port = "3133"
	}
	println("starting cors proxy on port " + port)
	http.ListenAndServe(":"+port, nil)
}

func cors(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin == "" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	} else {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Vary", "Origin")
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Cookie")

	reqHeaders := r.Header.Get("Access-Control-Request-Headers")
	if reqHeaders != "" {
		w.Header().Set("Access-Control-Allow-Headers", reqHeaders+", X-Cookie")
	} else {
		w.Header().Set("Access-Control-Allow-Headers", "*, X-Cookie")
	}

	w.Header().Set("Access-Control-Max-Age", "86400")
}
