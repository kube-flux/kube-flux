package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)
func main() {
	if len(os.Args) <= 1 {
		fmt.Println("Args Missing")
		panic("File path might be missing\n")
	}
	filename := os.Args[1]
	f, _ := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)

	h1 := func(a http.ResponseWriter, b *http.Request) {
		name := b.URL.Path[1:]
		_, _ = fmt.Fprintf(a, "Hello, %s!", name)
		fmt.Print("Writing into file\n")
		_, _ = f.WriteString("\n" + name)
	}

	h2 := func(a http.ResponseWriter, b *http.Request) {
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			fmt.Println("File reading error", err)
			return
		}
		_, _ = fmt.Fprintf(a, "List is: %s\n", string(data))
	}
	http.HandleFunc("/", h1)
	http.HandleFunc("/list", h2)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
