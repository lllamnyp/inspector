package handler

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"

	"github.com/lllamnyp/inspector/pkg/url"
)

func LogAndRewrite(entry url.URL) func(http.HandlerFunc) http.HandlerFunc {
	return func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			dump, _ := httputil.DumpRequest(r, true)
			os.Stdout.Write(dump)
			r2 := r.Clone(context.TODO())
			r2.URL.Path = strings.TrimPrefix(r2.URL.Path, entry.Path)
			r2.URL.RawPath = strings.TrimPrefix(r2.URL.RawPath, entry.Path)
			w2 := NewResponseWriterWrapper(w)
			w.Header()
			defer func() {
				fmt.Printf("%s", w2.String())
			}()
			h(w2, r2)
		}
	}
}

type ResponseWriterWrapper struct {
	w          *http.ResponseWriter
	body       *bytes.Buffer
	statusCode *int
}

// NewResponseWriterWrapper static function creates a wrapper for the http.ResponseWriter
func NewResponseWriterWrapper(w http.ResponseWriter) ResponseWriterWrapper {
	var buf bytes.Buffer
	var statusCode int = 200
	return ResponseWriterWrapper{
		w:          &w,
		body:       &buf,
		statusCode: &statusCode,
	}
}

func (rww ResponseWriterWrapper) Write(buf []byte) (int, error) {
	rww.body.Write(buf)
	return (*rww.w).Write(buf)
}

// Header function overwrites the http.ResponseWriter Header() function
func (rww ResponseWriterWrapper) Header() http.Header {
	return (*rww.w).Header()

}

// WriteHeader function overwrites the http.ResponseWriter WriteHeader() function
func (rww ResponseWriterWrapper) WriteHeader(statusCode int) {
	(*rww.statusCode) = statusCode
	(*rww.w).WriteHeader(statusCode)
}

func (rww ResponseWriterWrapper) String() string {
	var buf bytes.Buffer

	buf.WriteString("Response:")

	buf.WriteString("Headers:")
	for k, v := range (*rww.w).Header() {
		buf.WriteString(fmt.Sprintf("%s: %v", k, v))
	}

	buf.WriteString(fmt.Sprintf(" Status Code: %d", *(rww.statusCode)))

	buf.WriteString("Body")
	buf.WriteString(rww.body.String())
	return buf.String()
}
