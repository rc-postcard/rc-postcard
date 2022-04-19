package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
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
	acceptsPhysicalMail, err := strconv.ParseBool(r.FormValue("acceptsPhysicalMail"))
	if err != nil {
		log.Println(err)
		http.Error(w, "Error parsing form. Make sure to speicfy field acceptsPhysicalMail", http.StatusNotFound)
		return
	}

	createAddressResponse, err := lobClient.CreateAddress(name, address1, address2, city, state, zip, user.Id)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error creating address", http.StatusInternalServerError)
		return
	}

	if err = postgresClient.insertUser(user.Id, createAddressResponse.AddressId, user.Name, user.Email, acceptsPhysicalMail); err != nil {
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
}

func getAddress(w http.ResponseWriter, r *http.Request) {
	if !verifyRoute(w, r, http.MethodGet, "/addresses") {
		return
	}

	var user *User = r.Context().Value(userContextKey).(*User)

	lobAddressId, acceptsPhysicalMail, _, err := postgresClient.getUserInfo(user.Id)
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

	resp, err := json.Marshal(&GetAddressResponse{
		Name:                getAddressResponse.Name,
		AddressLine1:        getAddressResponse.AddressLine1,
		AddressLine2:        getAddressResponse.AddressLine2,
		AddressCity:         getAddressResponse.AddressCity,
		AddressState:        getAddressResponse.AddressState,
		AddressZip:          getAddressResponse.AddressZip,
		AddressCountry:      getAddressResponse.AddressCountry,
		AcceptsPhysicalMail: acceptsPhysicalMail,
	})

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
