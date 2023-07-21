package main

import (
	"flag"
	"fmt"
	"github.com/1f349/site-hosting/upload"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	listenFlag  string
	storageFlag string
)

func main() {
	flag.StringVar(&listenFlag, "listen", "", "Address to listen on")
	flag.StringVar(&storageFlag, "storage", "", "Path site files are stored in")
	if listenFlag == "" {
		log.Fatal("[SiteHosting] Missing listen flag")
	}
	if storageFlag == "" {
		log.Fatal("[SiteHosting] Missing storage flag")
	}
	_, err := os.Stat(storageFlag)
	if err != nil {
		log.Fatal("[SiteHosting] Failed to stat storage path, does the directory exist? Error: ", err)
	}

	uploadHandler := upload.New(storageFlag)

	router := httprouter.New()
	router.POST("/u/:site", uploadHandler.Handle)

	srv := &http.Server{
		Addr:    listenFlag,
		Handler: router,
	}
	log.Printf("[SiteHosting] Starting server on: '%s'\n", srv.Addr)
	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			if err == http.ErrServerClosed {
				log.Printf("[SiteHosting] The http server shutdown successfully\n")
			} else {
				log.Printf("[SiteHosting] Error trying to host the http server: %s\n", err.Error())
			}
		}
	}()

	scReload := make(chan os.Signal, 1)
	signal.Notify(scReload, syscall.SIGHUP)

	// Wait for exit signal
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	fmt.Println()

	// Stop server
	log.Printf("[SiteHosting] Stopping...")
	n := time.Now()

	// close http server
	_ = srv.Close()

	log.Printf("[SiteHosting] Took '%s' to shutdown\n", time.Now().Sub(n))
	log.Println("[SiteHosting] Goodbye")
}
