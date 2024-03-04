package main

// import (
// 	"net/http"
// )

type LoggingInfo struct {
	ipaddr string
	method string
	uri string
	code int
}

// func (app *application) addLoggging(h http.Handler) http.Handler {
// 	fn := func (w http.ResponseWriter, r *http.Request) {
// 		info := &LoggingInfo {
// 			method: r.Method,
// 			uri: r.URL.String(),
// 			ipaddr: r.RemoteAddr,
// 			}	
// 			app.logger.Printf("%s \"%s %s\"\n", info.ipaddr, info.method, info.uri)
// 			h.ServeHTTP(w, r)
// 	}
// 	return http.HandlerFunc(fn)
// }