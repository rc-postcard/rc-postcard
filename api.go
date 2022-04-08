package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

// pacCache is a personal access token cache used by the /tile API
var pacCache = map[string]*User{}

type Contact struct {
	RecurseId int    `json:"recurseId"`
	Name      string `json:"name"`
	Email     string `json:"email"`
}

type ContactsResponse struct {
	Contacts []*Contact `json:"contacts"`
}

func serveContacts(w http.ResponseWriter, r *http.Request) {
	if !verifyRoute(w, r, http.MethodGet, "/contacts") {
		return
	}

	contacts, err := postgresClient.getContacts()
	if err != nil {
		log.Println(err)
		http.Error(w, "Error getting contacts from db", http.StatusInternalServerError)
		return
	}

	resp, err := JSONMarshal(ContactsResponse{Contacts: contacts})
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

	createAddressResponse, err := lobClient.CreateAddress(name, address1, address2, city, state, zip)
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

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
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

func servePostcard(w http.ResponseWriter, r *http.Request) {
	if !verifyRoute(w, r, http.MethodPost, "/postcards") {
		return
	}

	query := r.URL.Query()
	isPreview, errIsPreview := strconv.ParseBool(query.Get("isPreview"))
	toRecurseId, errToRecurseId := strconv.Atoi(query.Get("toRecurseId"))
	if errIsPreview != nil || errToRecurseId != nil {
		log.Printf("Missing or malformed query parameter %v %v\n", errIsPreview, errToRecurseId)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Parse our multipart form, 10 << 20 specifies a maximum upload of 10 MB files.
	r.ParseMultipartForm(10 << 20)
	file, _, err := r.FormFile("front-postcard-file")
	if err != nil {
		log.Printf("Error Retrieving the File: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// read all of the contents of our uploaded file into a byte array
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	rcAddressId, err := postgresClient.getLobAddressId(recurseCenterRecurseId)
	if err != nil {
		log.Printf("Error getting recurse address: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var recipientAddressId string
	if !isPreview {
		recipientAddressId, err = postgresClient.getLobAddressId(toRecurseId)
		if err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	} else {
		recipientAddressId = rcAddressId
	}

	createPostCardResponse, lobError := lobClient.CreatePostCard(rcAddressId, recipientAddressId, fileBytes, isPreview)
	var resp []byte
	statusCodeCategory := lobError.StatusCode / 100
	if lobError.Err != nil || statusCodeCategory >= 5 {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	} else if statusCodeCategory == 3 || statusCodeCategory == 4 {
		resp, err = JSONMarshal(lobError)
		if err != nil {
			log.Println(err)
			w.WriteHeader(lobError.StatusCode)
			w.Header().Set("Content-Type", "application/json")
			w.Write(resp)
			return
		}
	}

	if !isPreview {
		createPostCardResponse.Url = ""
	}
	resp, err = JSONMarshal(createPostCardResponse)

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

func JSONMarshal(t interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(t)
	return buffer.Bytes(), err
}
