package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	lob "github.com/rc-postcard/rc-postcard/lob"
)

var client *http.Client = &http.Client{
	Timeout: time.Second * 20,
}

var lobClient *lob.Lob

var addr = flag.String("addr", ":8080", "http service address")

func main() {
	flag.Parse()

	// check oauth environment variables
	abort := false
	for _, env := range []string{
		"OAUTH_REDIRECT",
		"OAUTH_CLIENT_ID",
		"OAUTH_CLIENT_SECRET",
		"LOB_API_TEST_KEY",
		"PG_DATABASE_URL",
		"STRIPE_WEBHOOK_SECRET",
	} {
		if _, ok := os.LookupEnv(env); !ok {
			log.Println("Required environment variable missing:", env)
			abort = true
		}
	}
	if abort {
		log.Println("Aborting")
		os.Exit(1)
	}

	// setup postgres connection
	if err := postgresClient.setupPostgresConnection(); err != nil {
		log.Println("Error setting up postgres:", err)
		os.Exit(1)
	}
	defer db.Close()

	lobClient = lob.NewLob(client)

	var staticFS = http.FS(staticFiles)
	fs := http.FileServer(staticFS)

	http.HandleFunc("/", serveHome)
	http.HandleFunc("/login", serveLogin)
	http.HandleFunc("/auth", serveAuth)
	http.Handle("/static/", fs)

	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		favicon.Execute(w, nil)
		return
	})

	http.Handle("/addresses", authMiddleware(http.HandlerFunc(serveAddress)))
	http.Handle("/postcards", authMiddleware(http.HandlerFunc(servePostcards)))
	http.Handle("/contacts", authMiddleware(http.HandlerFunc(serveContacts)))
	http.Handle("/profiles", authMiddleware(http.HandlerFunc(serveProfiles)))
	http.HandleFunc("/stripeWebhook", serveStripeWebhook)

	log.Printf("Running on port %s\n", *addr)

	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
