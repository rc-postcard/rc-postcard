package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	lob "github.com/rc-postcard/rc-postcard/lob"
)

const (
	DigitalPreview string = "digital_preview"
	DigitalSend    string = "digital_send"
	PhysicalSend   string = "physical_send"
)

var validSendPostcardModes = []string{DigitalPreview, DigitalSend, PhysicalSend}

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
	if (!contains(validSendPostcardModes, mode)) || errToRecurseId != nil {
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

	if mode == PhysicalSend {
		// verify credits
		numCredits, err := postgresClient.getCredits(user.Id)
		if err != nil {
			log.Printf("Error getting user credits: %v\n", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if numCredits <= 0 {
			log.Printf("Credits less than or equal 0 or there was an error: %v\n", err)
			http.Error(w, "Credits error", http.StatusPaymentRequired)
			return
		}
	}

	fromAddress := lob.LobAddress{
		Name:         user.Name,
		AddressLine1: lob.RecurseAddressLine1,
		AddressLine2: lob.RecurseAddressLine2,
		AddressCity:  lob.RecurseAddressCity,
		AddressState: lob.RecurseAddressState,
		AddressZip:   lob.RecurseAddressZip,
	}

	var toAddress lob.LobAddress
	var useProductionKey bool
	if mode == PhysicalSend {
		// get sendee info
		receipientAddressId, recipientAcceptsPhysicalMail, _, _, err := postgresClient.getUserInfo(recurseCenterRecurseId)
		if err != nil {
			log.Printf("Error getting recurse address: %v\n", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if !recipientAcceptsPhysicalMail {
			log.Printf("Recipient does not accept physical mail %v\n", err)
			http.Error(w, "Recipient does not accept physical mail", http.StatusBadRequest)
			return
		}

		toAddress = lob.LobAddress{AddressId: receipientAddressId}
		useProductionKey = true
	} else if mode == DigitalSend {
		// get sendee info
		_, _, _, userName, err := postgresClient.getUserInfo(toRecurseId)
		if err != nil {
			log.Printf("Error getting user: %v\n", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		toAddress = lob.LobAddress{
			Name:         userName,
			AddressLine1: lob.RecurseAddressLine1,
			AddressLine2: lob.RecurseAddressLine2,
			AddressCity:  lob.RecurseAddressCity,
			AddressState: lob.RecurseAddressState,
			AddressZip:   lob.RecurseAddressZip,
		}
	} else {
		toAddress = lob.LobAddress{
			Name:         "Recurse Center",
			AddressLine1: lob.RecurseAddressLine1,
			AddressLine2: lob.RecurseAddressLine2,
			AddressCity:  lob.RecurseAddressCity,
			AddressState: lob.RecurseAddressState,
			AddressZip:   lob.RecurseAddressZip,
		}
	}

	lobCreatePostcardResponse, lobError := lobClient.CreatePostCard(toAddress, fromAddress, fileBytes, backTpl.String(), useProductionKey, user.Id, toRecurseId)
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

	if mode == DigitalPreview {
		createPostcardResponse.Url = lobCreatePostcardResponse.Url
	}

	if mode == PhysicalSend {
		err = postgresClient.decrementCredits(user.Id)
		if err != nil {
			log.Println(err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		numCredits, err := postgresClient.getCredits(user.Id)
		if err != nil {
			log.Printf("Error getting user credits: %v\n", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		createPostcardResponse.Credits = numCredits
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

func contains(arr []string, s string) bool {
	for _, elem := range arr {
		if elem == s {
			return true
		}
	}

	return false
}
