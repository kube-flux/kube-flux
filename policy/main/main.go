package main

import (
	"log"
	"net/http"

	"github.com/kube-flux/kube-flux/policy"
)

func main() {
	var handler, err = policy.NewPolicyHandler()
	if err != nil {
		log.Fatalln("Failed to initialize handler", "err:", err)
	}
	log.Println("Starting server")
	if err := http.ListenAndServe(":9999", handler); err != nil {
		log.Fatalln("Failed to start server", "err:", err)
	}
}
