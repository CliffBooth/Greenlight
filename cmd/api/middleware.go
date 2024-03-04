package main

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type LoggingInfo struct {
	ipaddr string
	method string
	uri string
	code int
}

func (app *application) logRequests(next http.Handler) http.Handler {
	fn := func (w http.ResponseWriter, r *http.Request) {
		defer next.ServeHTTP(w, r)

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

func (app *application) rateLimit(next http.Handler) http.Handler {
	if !app.config.limiter.enabled {
		return next
	}

	type client struct {
		limiter *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu sync.Mutex
		clients = make(map[string]*client)
	)

	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, client := range clients {
				if time.Since(client.lastSeen) >= 3 * time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()


	fn := func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		mu.Lock()

		if _, ok := clients[ip]; !ok {
			clients[ip] = &client{
				limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst),
			}
		}
		client := clients[ip]
		client.lastSeen = time.Now()
		if !client.limiter.Allow() {

			mu.Unlock()

			app.rateLimitExceededResponse(w, r)
			return
		}

		mu.Unlock()

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)

}