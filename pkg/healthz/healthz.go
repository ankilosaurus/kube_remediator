package healthz

import (
	"fmt"
	"net/http"

	httpmux "github.com/google/cadvisor/http/mux"
)

//TODO: add more checks as needed
func handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
	fmt.Print("alive")
}

func RegisterHandler(mux httpmux.Mux) error {
	mux.HandleFunc("/healthz", handleHealthz)
	return nil
}
