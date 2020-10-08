package policy

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/boltdb/bolt"
)

const (
	policyDB      = "policy.db"
	policyBucket  = "policyBucket"
	defaultPolicy = Black
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
		now := time.Now()
		if err = bucket.Put([]byte("CreatedAt"), []byte(now.String())); err != nil {
			log.Println("func", "NewPolicyHandler", "Failed to put default CreatedAt", "err:", err)
			return err
		}
		if err = bucket.Put([]byte("UpdatedAt"), []byte(now.String())); err != nil {
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
		}
	} else if r.Method == "GET" {
		// Get the values from database and write them to response
		handler.db.View(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte(policyBucket))
			if bucket == nil {
				err := errors.New("policy bucket doesn't exist")
				log.Println("func", "ServeHTTP", "Failed to write response to writer", "err:", err)
				return err
			}

			policy := Policy{
				Status:    Status(bucket.Get([]byte("Status"))),
				CreatedAt: string(bucket.Get([]byte("CreatedAt"))),
				UpdatedAt: string(bucket.Get([]byte("UpdatedAt"))),
			}

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(policy); err != nil {
				log.Println("func", "ServeHTTP", "Failed to write policy to writer", "err:", err)
				return err
			}
			log.Println("func", "ServeHTTP", "Response written")
			return nil
		})
	} else if r.Method == "PUT" {
		fmt.Fprintf(w, "Hello GET")
	}
}
