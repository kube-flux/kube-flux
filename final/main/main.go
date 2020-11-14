package main

import (
	"net/http"

	"github.com/kube-flux/kube-flux/final"
)

func main() {
	http.HandleFunc("/policy", final.Backend)
	// curl -X POST -H "Content-Type: application/json" -d '{"Red": {"TOP": 1, "Medium": 2, "LOW": 3}, "Yellow": {"TOP": 4, "Medium": 5, "LOW": 6}, "Green": {"TOP": 7, "Medium": 8, "LOW": 9}}' http://localhost:8888/factor
	http.HandleFunc("/factor", final.Factor)
	http.ListenAndServe(":8888", nil)
}
