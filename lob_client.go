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
	"os"
	"strconv"
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

type CreateAddressRequestMetadata struct {
	RCId string `json:"rc_id"`
}

type CreateAddressRequest struct {
	Name         string                       `json:"name"`
	AddressLine1 string                       `json:"address_line1"`
	AddressLine2 string                       `json:"address_line2"`
	AddressCity  string                       `json:"address_city"`
	AddressState string                       `json:"address_state"`
	AddressZip   string                       `json:"address_zip"`
	Metadata     CreateAddressRequestMetadata `json:"metadata"`
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

type LobError struct {
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
	Code       string `json:"code"`
	Err        error  `json:"err"`
}

type LobErrorResponse struct {
	LobError LobError `json:"error"`
}

type CreatePostcardMetadata struct {
	ToRcId   string `json:"to_rc_id"`
	FromRcId string `json:"from_rc_id"`
}

type GetPostcardsResponse struct {
	Data []Postcard `json:"data"`
}

type Metadata struct {
	ToRcId   string `json:"to_rc_id"`
	FromRcId string `json:"from_rc_id"`
}

type Postcard struct {
	Id       string   `json:"id"`
	Url      string   `json:"url"`
	Metadata Metadata `json:"metadata"`
}

func (*LobClient) GetPostcards(recipientRecurseId int) (*GetPostcardsResponse, error) {
	getPostcardsUrl := fmt.Sprintf("%s/%s/%s?metadata[to_rc_id]=%d", lobAddressBaseUrl, lobVersion, postcardsRoute, recipientRecurseId)
	req, err := http.NewRequest("GET", getPostcardsUrl, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(os.Getenv("LOB_API_TEST_KEY")+":")))

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var getPostcardsResponse GetPostcardsResponse
	if err := json.NewDecoder(resp.Body).Decode(&getPostcardsResponse); err != nil {
		log.Println(err)
		return nil, err
	}

	return &getPostcardsResponse, nil
}

func (*LobClient) GetAddress(lobAddressId string) (*GetAddressResponse, error) {
	getAddressUrl := fmt.Sprintf("%s/%s/%s/%s", lobAddressBaseUrl, lobVersion, addressesRoute, lobAddressId)
	req, err := http.NewRequest("GET", getAddressUrl, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(os.Getenv("LOB_API_TEST_KEY")+":")))

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
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
	deleteAddressUrl := fmt.Sprintf("%s/%s/%s/%s", lobAddressBaseUrl, lobVersion, addressesRoute, lobAddressId)
	req, err := http.NewRequest("DELETE", deleteAddressUrl, nil)
	if err != nil {
		log.Println(err)
		return err
	}
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(os.Getenv("LOB_API_TEST_KEY")+":")))

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return err
	}

	var deleteAddressResponse DeleteAddressResponse
	if err := json.NewDecoder(resp.Body).Decode(&deleteAddressResponse); err != nil {
		log.Println(err)
		return err
	}
	log.Println(deleteAddressResponse)

	return nil
}

func (*LobClient) CreateAddress(name, addressLine1, addressLine2, city, state, zipCode string, rcId int) (*CreateAddressResponse, error) {
	createAddressRequest := &CreateAddressRequest{
		Name:         name,
		AddressLine1: addressLine1,
		AddressLine2: addressLine2,
		AddressCity:  city,
		AddressState: state,
		AddressZip:   zipCode,
		Metadata:     CreateAddressRequestMetadata{RCId: strconv.Itoa(rcId)},
	}

	marshalledCreateAddressRequest, err := json.Marshal(createAddressRequest)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	createAddressUrl := fmt.Sprintf("%s/%s/%s", lobAddressBaseUrl, lobVersion, addressesRoute)
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
func (*LobClient) CreatePostCard(fromLobAddressId, toLobAddressId string, frontImage []byte, back string, isPreview bool, fromRcId, toRcId int) (*CreatePostcardResponse, *LobError) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	frontPart, _ := writer.CreateFormFile("front", "user-upload")
	io.Copy(frontPart, bytes.NewReader(frontImage))

	back = fmt.Sprintf("<body>%s</body", back)

	_ = writer.WriteField("back", back)
	_ = writer.WriteField("to", toLobAddressId)
	_ = writer.WriteField("from", fromLobAddressId)

	postcardMetadata, err := json.Marshal(CreatePostcardMetadata{FromRcId: strconv.Itoa(fromRcId), ToRcId: strconv.Itoa(toRcId)})
	if err != nil {
		log.Println(err)
		return nil, &LobError{Err: err}
	}

	metadataPart, err := writer.CreateFormField("metadata")
	if err != nil {
		log.Println(err)
		return nil, &LobError{Err: err}
	}

	_, err = metadataPart.Write(postcardMetadata)
	if err != nil {
		log.Println(err)
		return nil, &LobError{Err: err}
	}

	writer.Close()

	postPostcardUrl := fmt.Sprintf("%s/%s/%s", lobAddressBaseUrl, lobVersion, postcardsRoute)
	req, err := http.NewRequest("POST", postPostcardUrl, body)
	if err != nil {
		return nil, &LobError{Err: err}
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())
	var authHeader string
	if isPreview {
		authHeader = fmt.Sprintf("Basic %s",
			base64.StdEncoding.EncodeToString(
				[]byte(fmt.Sprintf("%s:", os.Getenv("LOB_API_TEST_KEY")))))
	} else {
		authHeader = fmt.Sprintf("Basic %s",
			base64.StdEncoding.EncodeToString(
				[]byte(fmt.Sprintf("%s:", os.Getenv("LOB_API_TEST_KEY")))))
	}
	req.Header.Set("Authorization", authHeader)

	resp, err := client.Do(req)
	if err != nil {
		return nil, &LobError{Err: err}
	}

	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		// update to read from actual response
		var createPostcardResponse CreatePostcardResponse
		if err := json.NewDecoder(resp.Body).Decode(&createPostcardResponse); err != nil {
			return nil, &LobError{Err: err}
		}
		return &createPostcardResponse, nil
	} else {
		var lobErrorResponse LobErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&lobErrorResponse); err != nil {
			return nil, &LobError{Err: err}
		}

		log.Println(lobErrorResponse)

		return nil, &lobErrorResponse.LobError
	}
}
