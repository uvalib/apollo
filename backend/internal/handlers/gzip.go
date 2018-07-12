package handlers

import (
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
)

// GzipResponseWriter is a Struct for manipulating io writer
type GzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (res GzipResponseWriter) Write(b []byte) (int, error) {
	if "" == res.Header().Get("Content-Type") {
		// If no content type, apply sniffing algorithm to un-gzipped body.
		res.Header().Set("Content-Type", http.DetectContentType(b))
	}
	return res.Writer.Write(b)
}

// GzipMiddleware will gzip responces
func GzipMiddleware(fn httprouter.Handle, force bool) httprouter.Handle {
	return func(res http.ResponseWriter, req *http.Request, pm httprouter.Params) {
		if !strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") && !force {
			fn(res, req, pm)
			return
		}
		res.Header().Set("Vary", "Accept-Encoding")
		res.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(res)
		defer gz.Close()
		log.Printf("Gzip API responce")
		gzr := GzipResponseWriter{Writer: gz, ResponseWriter: res}
		fn(gzr, req, pm)
	}
}
