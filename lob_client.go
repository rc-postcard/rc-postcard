package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"os"
)

type LobClient struct {
}

var lobClient = &LobClient{}

const lobAddressBaseUrl = "https://api.lob.com"
const lobVersion = "v1"
const addressesRoute = "addresses"
const postcardsRoute = "postcards"

type GetAddressResponse struct {
	Name           string `json:"name"`
	AddressLine1   string `json:"address_line1"`
	AddressLine2   string `json:"address_line2"`
	AddressCity    string `json:"address_city"`
	AddressState   string `json:"address_state"`
	AddressZip     string `json:"address_zip"`
	AddressCountry string `json:"address_country"`
}

type DeleteAddressResponse struct {
	AddressId string `json:"id"`
	Deleted   bool   `json:"deleted"`
}

type CreateAddressRequest struct {
	Name         string `json:"name"`
	AddressLine1 string `json:"address_line1"`
	AddressLine2 string `json:"address_line2"`
	AddressCity  string `json:"address_city"`
	AddressState string `json:"address_state"`
	AddressZip   string `json:"address_zip"`
}

type CreateAddressResponse struct {
	AddressId    string `json:"id"`
	Name         string `json:"name"`
	AddressLine1 string `json:"address_line1"`
	AddressLine2 string `json:"address_line2"`
	AddressCity  string `json:"address_city"`
	AddressState string `json:"address_state"`
	AddressZip   string `json:"address_zip"`
}

type CreatePostcardResponse struct {
	Url string `json:"url"`
}

func (*LobClient) GetAddress(lobAddressId string) (*GetAddressResponse, error) {
	getAddressUrl := lobAddressBaseUrl + "/" + lobVersion + "/" + addressesRoute + "/" + lobAddressId
	req, err := http.NewRequest("GET", getAddressUrl, nil)
	if err != nil {
		log.Println(err)
		panic(err)
	}

	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(os.Getenv("LOB_API_TEST_KEY")+":")))

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	// read body
	defer resp.Body.Close()
	var getAddressResponse GetAddressResponse
	if err := json.NewDecoder(resp.Body).Decode(&getAddressResponse); err != nil {
		return nil, err
	}
	return &getAddressResponse, nil
}

func (*LobClient) DeleteAddress(lobAddressId string) error {
	deleteAddressUrl := lobAddressBaseUrl + "/" + lobVersion + "/" + addressesRoute + "/" + lobAddressId
	req, err := http.NewRequest("DELETE", deleteAddressUrl, nil)
	if err != nil {
		log.Println(err)
		return err
	}
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(os.Getenv("LOB_API_TEST_KEY")+":")))

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return err
	}

	var deleteAddressResponse DeleteAddressResponse
	if err := json.NewDecoder(resp.Body).Decode(&deleteAddressResponse); err != nil {
		log.Println(err)
		return err
	}
	fmt.Println(deleteAddressResponse)

	return nil
}

func (*LobClient) CreateAddress(name, addressLine1, addressLine2, city, state, zipCode string) (*CreateAddressResponse, error) {
	createAddressRequest := &CreateAddressRequest{
		Name:         name,
		AddressLine1: addressLine1,
		AddressLine2: addressLine2,
		AddressCity:  city,
		AddressState: state,
		AddressZip:   zipCode,
	}

	marshalledCreateAddressRequest, err := json.Marshal(createAddressRequest)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	createAddressUrl := lobAddressBaseUrl + "/" + lobVersion + "/" + addressesRoute
	req, err := http.NewRequest("POST", createAddressUrl, bytes.NewBuffer(marshalledCreateAddressRequest))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(os.Getenv("LOB_API_TEST_KEY")+":")))

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var createAddressResponse CreateAddressResponse
	if err := json.NewDecoder(resp.Body).Decode(&createAddressResponse); err != nil {
		log.Println(err)
		return nil, err
	}

	return &createAddressResponse, nil
}

// https://gist.github.com/andrewmilson/19185aab2347f6ad29f5
// https://gist.github.com/mattetti/5914158/f4d1393d83ebedc682a3c8e7bdc6b49670083b84
func (*LobClient) CreatePostCard(fromLobAddressId, toLobAddressId string, frontImage []byte) (*CreatePostcardResponse, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// TODO come up with correct fileName
	frontPart, _ := writer.CreateFormFile("front", "user-upload")
	io.Copy(frontPart, bytes.NewReader(frontImage))

	_ = writer.WriteField("back", "<body>hello, back!</body>")
	_ = writer.WriteField("to", toLobAddressId)
	_ = writer.WriteField("from", fromLobAddressId)

	writer.Close()

	// TODO update with formatting
	postPostcardUrl := lobAddressBaseUrl + "/" + lobVersion + "/" + postcardsRoute
	req, err := http.NewRequest("POST", postPostcardUrl, body)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(os.Getenv("LOB_API_TEST_KEY")+":")))
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer resp.Body.Close()

	// update to read from actual response
	var createPostcardResponse CreatePostcardResponse
	if err := json.NewDecoder(resp.Body).Decode(&createPostcardResponse); err != nil {
		log.Println(err)
		return nil, err
	}

	// TODO pull
	fmt.Println(resp.StatusCode)
	b, err := httputil.DumpResponse(resp, true)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(string(b))

	return &createPostcardResponse, nil
}
