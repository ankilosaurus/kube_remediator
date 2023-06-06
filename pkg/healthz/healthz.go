package healthz

import (
	httpmux "github.com/google/cadvisor/http/mux"
	"net/http"
)

// TODO: add more checks as needed
func handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func RegisterHandler(mux httpmux.Mux) error {
	mux.HandleFunc("/healthz", handleHealthz)
	return nil
}
