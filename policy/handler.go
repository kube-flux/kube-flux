package policy

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/boltdb/bolt"
)

const (
	policyDB      = "policy.db"
	policyBucket  = "policyBucket"
	defaultPolicy = "Black"
)

type policyHandler struct {
	db *bolt.DB
}

func NewPolicyHandler() (*policyHandler, error) {
	log.Println("func", "NewPolicyHandler", "Opening db")
	db, err := bolt.Open(policyDB, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Println("func", "NewPolicyHandler", "Failed to initialize db", "err:", err)
		return nil, err
	}

	// Initial Policy in database
	err = db.Update(func(tx *bolt.Tx) error {
		log.Println("func", "NewPolicyHandler", "Creating bucket")
		bucket, err := tx.CreateBucketIfNotExists([]byte(policyBucket))
		if err != nil {
			log.Println("func", "NewPolicyHandler", "Failed to create bucket", "err:", err)
			return err
		}
		log.Println("func", "NewPolicyHandler", "Created bucket")

		if err = bucket.Put([]byte("Status"), []byte(defaultPolicy)); err != nil {
			log.Println("func", "NewPolicyHandler", "Failed to put default status", "err:", err)
			return err
		}
		if err = bucket.Put([]byte("CreatedAt"), []byte(defaultPolicy)); err != nil {
			log.Println("func", "NewPolicyHandler", "Failed to put default CreatedAt", "err:", err)
			return err
		}
		if err = bucket.Put([]byte("UpdatedAt"), []byte(defaultPolicy)); err != nil {
			log.Println("func", "NewPolicyHandler", "Failed to put default UpdatedAt", "err:", err)
			return err
		}
		log.Println("func", "NewPolicyHandler", "Added default parameters")
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &policyHandler{db: db}, nil
}

func (handler *policyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("func", "ServeHTTP", "Start handling policy request")
	defer log.Println("func", "ServeHTTP", "Finished handling policy request")

	// No supports for POST & DELETE
	if r.Method == "POST" || r.Method == "DELETE" {
		log.Println("func", "ServeHTTP", "No support for POST & DELETE")
		if _, err := fmt.Fprintf(w, "No support for POST & DELETE"); err != nil {
			log.Println("func", "ServeHTTP", "Failed to write to writer", "err:", err)
			return
		}
	}

	if r.Method == "GET" {
		fmt.Fprintf(w, "Hello GET")
	}

	if r.Method == "PUT" {
		fmt.Fprintf(w, "Hello GET")
	}
}
