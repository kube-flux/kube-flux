package final

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

// ----------------- policy -----------------

type Policy struct {
	Status string
	Factor map[string]map[string]int32
}

var policy *Policy

func init() {
	green := map[string]int32{"Top": 4, "Medium": 4, "Low": 4}
	yellow := map[string]int32{"Top": 2, "Medium": 2, "Low": 2}
	red := map[string]int32{"Top": 1, "Medium": 1, "Low": 1}

	initFactor := map[string]map[string]int32{"Green": green, "Yellow": yellow, "Red": red}

	policy = &Policy{
		Status: "Green",
		Factor: initFactor,
	}
}

// backend handles policy & importance factor
func Backend(w http.ResponseWriter, req *http.Request) {
	// Setup response
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS, PUT")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	if req.Method == "OPTIONS" {
		return
	}

	if req.Method == "GET" {
		log.Println("func", "ServeHTTP", "Handling GET request /policy")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(policy)

		log.Println("func", "ServeHTTP", "Handled GET request for /policy")
		return
	}

	if req.Method == "PUT" {
		log.Println("func", "ServeHTTP", "Handling PUT request")
		w.WriteHeader(http.StatusNoContent)

		log.Println("func", "NewPolicyHandler", "Decoding request body")
		decoder := json.NewDecoder(req.Body)
		var request Policy
		if err := decoder.Decode(&request); err != nil {
			err := errors.New("failed to decode to Policy")
			log.Println("func", "ServeHTTP", "decode policy from request err:", err)
		}

		if policy.Status == request.Status {
			log.Println("func", "ServeHTTP", "Same status")
			return
		}

		policy.Status = request.Status
		executePolicy()
		return
	}
}

func Factor(w http.ResponseWriter, req *http.Request) {
	// Setup response
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS, PUT")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	if req.Method == "OPTIONS" {
		return
	}

	decoder := json.NewDecoder(req.Body)
	var factor = make(map[string]map[string]int32)
	decoder.Decode(&factor)
	setFactor(factor)
	log.Printf("Policy: %v\n", policy)

}

func setFactor(factor map[string]map[string]int32) {
	policy.Factor = factor
	executePolicy()
}

func executePolicy() {
	// TODO
}
