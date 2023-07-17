package httputil

import "net/http"

// CorsHandler wraps an http Handler and sets the CORS headers on the response.
type CorsHandler struct {
	Handler http.Handler
}

func WriteCorsHeaders(header http.Header) {
	header.Set("Access-Control-Allow-Origin", "*")
	header.Set("Access-Control-Allow-Methods", "*")
	header.Set("Access-Control-Allow-Headers", "content-type, authorization")
}

func (h *CorsHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	WriteCorsHeaders(res.Header())

	h.Handler.ServeHTTP(res, req)
}
