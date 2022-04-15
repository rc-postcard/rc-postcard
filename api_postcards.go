package main

import (
	"bytes"
	"encoding/json"
	_ "image/png"
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

func sendPostcards(w http.ResponseWriter, r *http.Request) {
	if !verifyRoute(w, r, http.MethodPost, "/postcards") {
		return
	}

	var user *User = r.Context().Value(userContextKey).(*User)

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

	fileBytes, err := ioutil.ReadAll(file)

	// // read all of the contents of our uploaded file into a byte array

	// // Decoding gives you an Image.
	// // If you have an io.Reader already, you can give that to Decode
	// // without reading it into a []byte.
	// image, _, err := image.Decode(file)
	// // check err
	// if err != nil {
	// 	log.Printf("Error decoding image: %v\n", err)
	// 	http.Error(w, "Bad Request", http.StatusBadRequest)
	// 	return
	// }

	// newImage := resize.Resize(
	// 	1875,
	// 	1275,
	// 	image,
	// 	resize.Lanczos3,
	// )

	// Encode uses a Writer, use a Buffer if you need the raw []byte
	// var fileBytes bytes.Buffer
	// err = jpeg.Encode(&fileBytes, newImage, nil)

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

	createPostCardResponse, lobError := lobClient.CreatePostCard(rcAddressId, recipientAddressId, fileBytes, backTpl.String(), isPreview, user.Id, toRecurseId)
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

	if !isPreview {
		createPostCardResponse.Url = ""
	}
	resp, err := JSONMarshal(createPostCardResponse)

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
