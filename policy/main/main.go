package main

import (
	"log"
	"net/http"

	"github.com/USF-VMware-EnergyAwareDatacenter-2020/kube-flux/policy"
)

func main() {
	handler := policy.NewPolicyHandler()
	if err := http.ListenAndServe(":9999", handler); err != nil {
		log.Fatalln("Failed to start server", "err", err)
	}
}
