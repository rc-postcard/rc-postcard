package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	lob "github.com/rc-postcard/rc-postcard/lob"
	"github.com/stripe/stripe-go"
)

// pacCache is a personal access token cache used by the /tile API
var pacCache = map[string]*User{}

type Contact struct {
	RecurseId           int    `json:"recurseId"`
	Name                string `json:"name"`
	Email               string `json:"email"`
	Batch               string `json:"batch"`
	AcceptsPhysicalMail bool   `json:"acceptsPhysicalMail"`
}

type ContactsResponse struct {
	Contacts []*Contact `json:"contacts"`
	Credits  int        `json:"credits"`
}

func serveContacts(w http.ResponseWriter, r *http.Request) {
	if !verifyRoute(w, r, http.MethodGet, "/contacts") {
		return
	}

	var user *User = r.Context().Value(userContextKey).(*User)

	contacts, err := postgresClient.getContacts()
	if err != nil {
		log.Println(err)
		http.Error(w, "Error getting contacts from db", http.StatusInternalServerError)
		return
	}

	credits, err := postgresClient.getCredits(user.Id)
	if err != nil {
		log.Println(err)
		credits = 0
	}

	resp, err := JSONMarshal(ContactsResponse{Contacts: contacts, Credits: credits})
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
	return

}

func serveAddress(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		createOrUpdateAddress(w, r)
	} else if r.Method == http.MethodGet {
		getAddress(w, r)
	} else {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
}

func handleCheckoutSessionCompleted(checkoutSession stripe.CheckoutSession) {
	recurseIdString := checkoutSession.ClientReferenceID
	recurseId, _ := strconv.Atoi(recurseIdString)
	postgresClient.incrementCredits(recurseId)
}

func serveStripeWebhook(w http.ResponseWriter, req *http.Request) {
	const MaxBodyBytes = int64(65536)
	req.Body = http.MaxBytesReader(w, req.Body, MaxBodyBytes)
	payload, err := ioutil.ReadAll(req.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading request body: %v\n", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	event := stripe.Event{}

	if err := json.Unmarshal(payload, &event); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse webhook body json: %v\n", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Unmarshal the event data into an appropriate struct depending on its Type
	switch event.Type {
	case "checkout.session.completed":
		log.Printf("Event %v\n", event)
		var checkoutSession stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &checkoutSession)
		log.Printf("Client reference id %v\n", checkoutSession.ClientReferenceID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// Then define and call a func to handle the successful payment intent.
		handleCheckoutSessionCompleted(checkoutSession)
	default:
		fmt.Fprintf(os.Stderr, "Unhandled event type: %s\n", event.Type)
	}

	w.WriteHeader(http.StatusOK)
}

func createOrUpdateAddress(w http.ResponseWriter, r *http.Request) {
	if !verifyRoute(w, r, http.MethodPost, "/addresses") {
		return
	}

	var user *User = r.Context().Value(userContextKey).(*User)

	if err := r.ParseForm(); err != nil {
		log.Println(err)
		http.Error(w, "Form error", http.StatusBadRequest)
		return
	}

	name, address1, address2 := r.FormValue("name"), r.FormValue("address1"), r.FormValue("address2")
	city, state, zip := r.FormValue("city"), r.FormValue("state"), r.FormValue("zip")
	acceptsPhysicalMail, err := strconv.ParseBool(r.FormValue("acceptsPhysicalMail"))
	if err != nil {
		log.Println(err)
		http.Error(w, "Error parsing form. Make sure to speicfy field acceptsPhysicalMail", http.StatusNotFound)
		return
	}

	if acceptsPhysicalMail {
		verifyAddressResponse, err := lobClient.VerifyAddressBySendingTestPostcard(address1, address2, city, state, zip)
		if err != nil {
			log.Println(err)
			http.Error(w, "Error verifying address", http.StatusBadRequest)
			return
		}

		if verifyAddressResponse.Deliverability == lob.Undeliverable {
			log.Println("Address undeliverable")
			http.Error(w, "Address undeliverable", http.StatusBadRequest)
			return
		}
	}

	createAddressResponse, err := lobClient.CreateAddress(name, address1, address2, city, state, zip, user.Id, true)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error creating address", http.StatusInternalServerError)
		return
	}

	if err = postgresClient.updateAddress(user.Id, createAddressResponse.AddressId, acceptsPhysicalMail); err != nil {
		log.Println(err)
		http.Error(w, "Error setting address in database", http.StatusInternalServerError)
		return
	}

	resp, err := JSONMarshal(new(struct{}))
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
	return
}

type GetAddressResponse struct {
	Name                string `json:"name"`
	AddressLine1        string `json:"address_line1"`
	AddressLine2        string `json:"address_line2"`
	AddressCity         string `json:"address_city"`
	AddressState        string `json:"address_state"`
	AddressZip          string `json:"address_zip"`
	AddressCountry      string `json:"address_country"`
	AcceptsPhysicalMail bool   `json:"acceptsPhysicalMail"`
	RecurseId           int    `json:"recurse_id"`
	Email               string `json:"email"`
}

func getAddress(w http.ResponseWriter, r *http.Request) {
	if !verifyRoute(w, r, http.MethodGet, "/addresses") {
		return
	}

	var user *User = r.Context().Value(userContextKey).(*User)

	lobAddressId, acceptsPhysicalMail, _, _, err := postgresClient.getUserInfo(user.Id)
	if err != nil {
		log.Println(err)
		http.Error(w, "No address found that corresponds to this user.", http.StatusNotFound)
		return
	}

	var getAddressResponse GetAddressResponse
	if lobAddressId != "" {
		lobAddressResponse, err := lobClient.GetAddress(lobAddressId, true)
		if err != nil {
			log.Println(err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		getAddressResponse = GetAddressResponse{
			Name:                lobAddressResponse.Name,
			AddressLine1:        lobAddressResponse.AddressLine1,
			AddressLine2:        lobAddressResponse.AddressLine2,
			AddressCity:         lobAddressResponse.AddressCity,
			AddressState:        lobAddressResponse.AddressState,
			AddressZip:          lobAddressResponse.AddressZip,
			AddressCountry:      lobAddressResponse.AddressCountry,
			AcceptsPhysicalMail: acceptsPhysicalMail,
			RecurseId:           user.Id,
			Email:               user.Email,
		}
	} else {
		getAddressResponse = GetAddressResponse{
			Name:                user.Name,
			AddressLine1:        lob.RecurseAddressLine1,
			AddressLine2:        lob.RecurseAddressLine2,
			AddressCity:         lob.RecurseAddressCity,
			AddressState:        lob.RecurseAddressState,
			AddressZip:          lob.RecurseAddressZip,
			AddressCountry:      lob.RecurseAddressCountry,
			AcceptsPhysicalMail: false,
			RecurseId:           user.Id,
			Email:               user.Email,
		}
	}

	resp, err := json.Marshal(&getAddressResponse)

	if err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
	return
}
