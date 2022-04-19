package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

func servePostcards(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		sendPostcards(w, r)
	} else if r.Method == http.MethodGet {
		getPostcards(w, r)
	} else {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
}

func getPostcards(w http.ResponseWriter, r *http.Request) {
	if !verifyRoute(w, r, http.MethodGet, "/postcards") {
		return
	}

	var user *User = r.Context().Value(userContextKey).(*User)

	postcards, err := lobClient.GetPostcards(user.Id)

	if err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	resp, err := JSONMarshal(postcards)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
	return
}

type CreatePostcardResponse struct {
	Url     string `json:"url"`
	Credits int    `json:"credits"`
}

func sendPostcards(w http.ResponseWriter, r *http.Request) {
	if !verifyRoute(w, r, http.MethodPost, "/postcards") {
		return
	}

	var user *User = r.Context().Value(userContextKey).(*User)

	query := r.URL.Query()
	mode := query.Get("mode")
	toRecurseId, errToRecurseId := strconv.Atoi(query.Get("toRecurseId"))
	// TODO formalize this
	if (mode != "digital_send" && mode != "digital_preview" && mode != "physical_send") || errToRecurseId != nil {
		log.Printf("Missing or malformed query parameter %s %v\n", mode, errToRecurseId)
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

	fileBytes, err := ioutil.ReadAll(file)

	// check err
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	back := r.FormValue("back")

	var backTpl bytes.Buffer
	if err = backOfPostcard.Execute(&backTpl, struct{ Message string }{Message: back}); err != nil {
		log.Println(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	rcAddressId, _, numCredits, err := postgresClient.getUserInfo(user.Id)
	if err != nil {
		log.Printf("Error getting recurse address: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if mode == "physical_send" && numCredits <= 0 {
		log.Printf("Credits less than or equal 0 or there was an error: %v\n", err)
		http.Error(w, "Credits error", http.StatusPaymentRequired)
		return
	}

	var recipientAddressId string
	if mode == "digital_send" || mode == "physical_send" {
		recipientAddressId, err = postgresClient.getLobAddressId(toRecurseId)
		if err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	} else {
		recipientAddressId = rcAddressId
	}

	lobCreatePostcardResponse, lobError := lobClient.CreatePostCard(rcAddressId, recipientAddressId, fileBytes, backTpl.String(), mode, user.Id, toRecurseId)
	if lobError != nil && (lobError.Err != nil || lobError.StatusCode/100 >= 5) {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	} else if lobError != nil && (lobError.StatusCode/100 == 3 || lobError.StatusCode/100 == 4) {
		resp, err := JSONMarshal(lobError)
		if err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}

		w.WriteHeader(lobError.StatusCode)
		w.Header().Set("Content-Type", "application/json")
		w.Write(resp)
		return
	}

	createPostcardResponse := &CreatePostcardResponse{Credits: 0}

	if mode == "digital_preview" {
		createPostcardResponse.Url = lobCreatePostcardResponse.Url
	}

	if mode == "physical_send" {
		// TODO get after set, potential race condition
		err = postgresClient.decrementCredits(user.Id)
		if err != nil {
			log.Println(err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		createPostcardResponse.Credits = numCredits - 1
	}

	resp, err := JSONMarshal(createPostcardResponse)

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
