package main

import (
	"encoding/json"
	"log"
	"net/http"
)

// pacCache is a personal access token cache used by the /tile API
var pacCache = map[string]*User{}

type Contact struct {
	RecurseId           int    `json:"recurseId"`
	Name                string `json:"name"`
	Email               string `json:"email"`
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
		createAddress(w, r)
	} else if r.Method == http.MethodDelete {
		deleteAddress(w, r)
	} else if r.Method == http.MethodGet {
		getAddress(w, r)
	} else {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
}

func deleteAddress(w http.ResponseWriter, r *http.Request) {
	if !verifyRoute(w, r, http.MethodDelete, "/addresses") {
		return
	}

	var user *User = r.Context().Value(userContextKey).(*User)

	lobAddressId, err := postgresClient.getLobAddressId(user.Id)
	if err != nil {
		log.Println(err)
		http.Error(w, "No address found that corresponds to this user.", http.StatusNotFound)
		return
	}

	if err := lobClient.DeleteAddress(lobAddressId); err != nil {
		log.Println(err)
		http.Error(w, "Error deleting address", http.StatusInternalServerError)
		return
	}

	if err = postgresClient.deleteUser(user.Id); err != nil {
		log.Println(err)
		http.Error(w, "Error setting address in database", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	return
}

func createAddress(w http.ResponseWriter, r *http.Request) {
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

	createAddressResponse, err := lobClient.CreateAddress(name, address1, address2, city, state, zip, user.Id)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error creating address", http.StatusInternalServerError)
		return
	}

	if err = postgresClient.insertUser(user.Id, createAddressResponse.AddressId, user.Name, user.Email); err != nil {
		log.Println(err)
		http.Error(w, "Error setting address in database", http.StatusInternalServerError)
		return
	}

	// Hide address id from user
	createAddressResponse.AddressId = ""
	resp, err := JSONMarshal(createAddressResponse)
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

func getAddress(w http.ResponseWriter, r *http.Request) {
	if !verifyRoute(w, r, http.MethodGet, "/addresses") {
		return
	}

	var user *User = r.Context().Value(userContextKey).(*User)

	lobAddressId, err := postgresClient.getLobAddressId(user.Id)
	if err != nil {
		log.Println(err)
		http.Error(w, "No address found that corresponds to this user.", http.StatusNotFound)
		return
	}

	// use lobClient to get address
	getAddressResponse, err := lobClient.GetAddress(lobAddressId)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(getAddressResponse)
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
