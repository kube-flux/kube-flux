package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/boltdb/bolt"
)

func main() {

	http.HandleFunc("/hello/v1/", nameHandler)
	http.HandleFunc("/hello/v1/list", listHandler)
	http.ListenAndServe(":8080", nil)
}

func nameHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintf(w, "Hello ")
	params := strings.Split(r.URL.Path, "/")
	user := params[len(params)-1]
	fmt.Fprintf(w, user)

	// Open db
	db, err := bolt.Open("my.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("MyBucket"))
		fmt.Println("Opened bucket")
		if err != nil {
			fmt.Println("Err bucket")
			return fmt.Errorf("create bucket: %s", err)
		}

		err = b.Put([]byte(user), []byte("1"))
		fmt.Print("Put user:")
		fmt.Println(user)
		return err
	})

}

func listHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "User List\n")

	db, err := bolt.Open("my.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("MyBucket"))

		c := b.Cursor()

		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			fmt.Fprintf(w, string(k))
			fmt.Fprintf(w, "\n")
		}

		return nil
	})
}
