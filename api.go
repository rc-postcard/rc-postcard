package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

// pacCache is a personal access token cache used by the /tile API
var pacCache = map[string]*User{}

func serveAddress(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		createAddress(w, r)
	} else {
		getAddress(w, r)
	}
}

func createAddress(w http.ResponseWriter, r *http.Request) {
	if !verifyRoute(w, r, http.MethodPost, "/address") {
		return
	}

	// Add auth

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
	if !verifyRoute(w, r, http.MethodGet, "/address") {
		return
	}

	// authenticate and get userId from token
	// TODO add support for session
	_, err := authPersonalAccessToken(r)
	if err != nil {
		log.Println(err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// get address using userId
	// TODO update to use postgres
	lobAddressId := os.Getenv("LOB_TEST_ADDRESS_ID")

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

func servePostcardPreview(w http.ResponseWriter, r *http.Request) {
	if !verifyRoute(w, r, http.MethodPost, "/postcardPreview") {
		return
	}

	// authenticate and get userId from token
	// TODO add support for session
	_, err := authPersonalAccessToken(r)
	if err != nil {
		log.Println(err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse our multipart form, 10 << 20 specifies a maximum
	// upload of 10 MB files.
	r.ParseMultipartForm(10 << 20)
	// FormFile returns the first file for the given key `myFile`
	// it also returns the FileHeader so we can get the Filename,
	// the Header and the size of the file
	file, _, err := r.FormFile("front-postcard-file")
	if err != nil {
		log.Println("Error Retrieving the File")
		log.Println(err)
		return
	}
	defer file.Close()

	// read all of the contents of our uploaded file into a
	// byte array
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Println(err)
	}

	createPostCardResponse, err := lobClient.CreatePostCard(os.Getenv("LOB_TEST_ADDRESS_ID"), os.Getenv("LOB_TEST_ADDRESS_ID"), fileBytes)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
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

// authPersonalAccessToken will authenticate an Authorization header by
// forwarding a request to recurse.com API and cache a successful result
// in pacCache.
func authPersonalAccessToken(r *http.Request) (*User, error) {
	// get token
	pacToken := r.Header.Get("Authorization")
	if pacToken == "" {
		return nil, errors.New("missing authentication token")
	}
	// check cache
	if u, ok := pacCache[pacToken]; ok {
		return u, nil
	}
	// send request to recurse.com
	req, err := http.NewRequest(http.MethodGet, "https://recurse.com/api/v1/profiles/me", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", pacToken)

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("unauthorized")
	}

	// read body
	defer resp.Body.Close()
	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	// update cache
	pacCache[pacToken] = &user
	return pacCache[pacToken], nil
}

func JSONMarshal(t interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(t)
	return buffer.Bytes(), err
}
