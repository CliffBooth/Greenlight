package main

import (
	"fmt"
	"net/http"
	"strings"
)

type LoggingInfo struct {
	ipaddr string
	method string
	uri string
	code int
}

func (app *application) logRequests(h http.Handler) http.Handler {
	fn := func (w http.ResponseWriter, r *http.Request) {
		defer h.ServeHTTP(w, r)

		info := &LoggingInfo {
			method: r.Method,
			uri: r.URL.String(),
			ipaddr: r.RemoteAddr,
			}	
			//if trailing slash redirection is turned on
			//in order to not log the same request twice we return.
			path := r.URL.Path
			if strings.HasSuffix(path, "/") { 
				return
			}
			app.logger.PrintInfo("request", map[string]string{
				"ip": info.ipaddr,
				"method": info.method,
				"uri": info.uri,
			})
	}
	return http.HandlerFunc(fn)
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
