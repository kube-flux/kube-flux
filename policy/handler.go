package policy

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/boltdb/bolt"
)

type policyHandler struct {
	db *bolt.DB
}

func NewPolicyHandler() *policyHandler {
	db, err := bolt.Open("policy.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatalln("Failed initialize db", "err", err)
	}
	return &policyHandler{db: db}
}

func (handler *policyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("Start handling policy request")
	defer log.Println("Finished handling policy request")

	// Detect Http request type
	if r.Method == "GET" {
		fmt.Fprintf(w, "Hello GET")
	}
}
