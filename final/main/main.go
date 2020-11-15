package main

import (
	"net/http"

	"github.com/kube-flux/kube-flux/final"
)

// curl -X PUT -H "Content-Type: application/json" -d '{"Red": {"TOP": 1, "Medium": 1, "LOW": 0}, "Yellow": {"TOP": 2, "Medium": 2, "LOW": 0}, "Green": {"TOP": 4, "Medium": 4, "LOW": 0}}' http://localhost:8888/factor
func main() {

	http.HandleFunc("/policy", final.Backend)
	http.HandleFunc("/factor", final.Factor)
	http.ListenAndServe(":8888", nil)

}
