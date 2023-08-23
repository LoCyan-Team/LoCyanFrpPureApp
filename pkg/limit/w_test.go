package limit

import (
	"fmt"
	"net/http"
	"testing"
)

func TestHttp(_ *testing.T) {
	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		lw := NewWriterWithLimit(w, 10*KB)
		for {
			fmt.Fprintf(lw, "x")
		}
	})

	// nolint:gosec // Don't need to set timeouts for tests
	http.ListenAndServe(":62542", nil)
}
