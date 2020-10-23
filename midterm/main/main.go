package main

import (
	"kube-flux/midterm"
	"log"
	"net/http"
	"os"
)


func main() {
	filePath := os.Args[1]                         //Pass .pem file as a command line argument
	clusterIP := os.Args[2]                        //Pass cluster IP address
	var handler, err = midterm.NewPolicyHandler(filePath, clusterIP)
	if err != nil {
		log.Fatalln("Failed to initialize handler", "err:", err)
	}
	log.Println("Starting server")
	if err := http.ListenAndServe(":9999", handler); err != nil {
		log.Fatalln("Failed to start server", "err:", err)
	}
}
