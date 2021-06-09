package main

import (
	"flag"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/handlers"
	log "github.com/sirupsen/logrus"

	"github.com/droidvirt/droidvirt-ctrl/backend/handler"
)

const (
	DefaultListenPort int    = 8080
	DefaultNamespace  string = "droidvirt"
)

var (
	listenPort   int
	crdNamespace string
)

func main() {
	flag.IntVar(&listenPort, "port", DefaultListenPort, `port this server listen to`)
	flag.StringVar(&crdNamespace, "namespace", DefaultNamespace, `DroidVirt and DroidVirtVolume search namespace`)
	flag.Parse()

	h, err := handler.NewAPIHandler(crdNamespace)
	if err != nil {
		log.Fatalf("Failed to create route handler: %v", err)
	}
	router := handler.NewRouter(h)

	http.Handle("/", router)

	p := ":" + strconv.Itoa(listenPort)
	log.Infof("droidvirt server listens on: %v", p)
	if err = http.ListenAndServe(p, handlers.LoggingHandler(os.Stdout, http.DefaultServeMux)); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
