package policy

import (
	"encoding/json"
	"errors"
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
		if err = bucket.Put([]byte("UpdatedAt"), []byte(time.Now().String())); err != nil {
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

	// Setup response
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS, PUT")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	if r.Method == "OPTIONS" {
		return
	}

	if r.Method == "GET" {
		log.Println("func", "ServeHTTP", "Handling GET request")
		err := handler.db.View(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte(policyBucket))
			if bucket == nil {
				err := errors.New("policy bucket doesn't exist")
				log.Println("func", "ServeHTTP", "Failed to find bucket", "err:", err)
				return err
			}

			policy := Policy{
				Status:    Status(bucket.Get([]byte("Status"))),
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
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if r.Method == "PUT" {
		log.Println("func", "ServeHTTP", "It's a PUT request")
		err := handler.db.Update(func(tx *bolt.Tx) error {
			// Get bucket
			log.Println("func", "NewPolicyHandler", "Getting bucket")
			bucket := tx.Bucket([]byte(policyBucket))
			if bucket == nil {
				err := errors.New("policy bucket doesn't exist")
				log.Println("func", "ServeHTTP", "Failed to find bucket", "err:", err)
				return err
			}

			// Get Status from request
			log.Println("func", "NewPolicyHandler", "Decoding request body")
			decoder := json.NewDecoder(r.Body)
			var policy Policy
			if err := decoder.Decode(&policy); err != nil {
				err := errors.New("Failed to decode request to Policy")
				log.Println("func", "ServeHTTP", "err:", err)
				return err
			}

			// Update Status
			log.Println("func", "NewPolicyHandler", "Updating Policy")
			if err := bucket.Put([]byte("Status"), []byte(policy.Status)); err != nil {
				log.Println("func", "NewPolicyHandler", "Failed to put status", "err:", err)
				return err
			}
			if err := bucket.Put([]byte("UpdatedAt"), []byte(policy.UpdatedAt)); err != nil {
				log.Println("func", "NewPolicyHandler", "Failed to put UpdatedAt", "err:", err)
				return err
			}

			w.WriteHeader(http.StatusNoContent)
			log.Println("func", "NewPolicyHandler", "Updated Policy", string(bucket.Get([]byte("Status"))), string(bucket.Get([]byte("UpdatedAt"))))
			return nil
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
}
