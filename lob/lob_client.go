package lob

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
	"time"
)

const (
	RecurseAddressLine1   = "397 Bridge Street"
	RecurseAddressLine2   = ""
	RecurseAddressCity    = "Brooklyn"
	RecurseAddressState   = "NY"
	RecurseAddressZip     = "11201"
	RecurseAddressCountry = "US"
)

type Lob struct {
	httpClient *http.Client
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

const lobAddressBaseUrl = "https://api.lob.com"
const lobVersion = "v1"
const addressesRoute = "addresses"
const postcardsRoute = "postcards"

func NewLob(httpClient *http.Client) *Lob {
	return &Lob{
		httpClient: httpClient,
	}
}

type LobGetPostcardsResponse struct {
	Data []struct {
		Id       string `json:"id"`
		Url      string `json:"url"`
		Metadata struct {
			ToRcId   string `json:"to_rc_id"`
			FromRcId string `json:"from_rc_id"`
		} `json:"metadata"`
		DateCreated time.Time `json:"date_created"`
	} `json:"data"`
}

func (l *Lob) GetPostcards(recipientRecurseId int) (*LobGetPostcardsResponse, error) {
	getPostcardsUrl := fmt.Sprintf("%s/%s/%s?metadata[to_rc_id]=%d", lobAddressBaseUrl, lobVersion, postcardsRoute, recipientRecurseId)
	req, err := http.NewRequest("GET", getPostcardsUrl, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(os.Getenv("LOB_API_TEST_KEY")+":")))

	resp, err := l.httpClient.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var getPostcardsResponse LobGetPostcardsResponse
	if err := json.NewDecoder(resp.Body).Decode(&getPostcardsResponse); err != nil {
		log.Println(err)
		return nil, err
	}

	return &getPostcardsResponse, nil
}

type LobGetAddressResponse struct {
	Name           string `json:"name"`
	AddressLine1   string `json:"address_line1"`
	AddressLine2   string `json:"address_line2"`
	AddressCity    string `json:"address_city"`
	AddressState   string `json:"address_state"`
	AddressZip     string `json:"address_zip"`
	AddressCountry string `json:"address_country"`
}

func (l *Lob) GetAddress(lobAddressId string, isLive bool) (*LobGetAddressResponse, error) {
	getAddressUrl := fmt.Sprintf("%s/%s/%s/%s", lobAddressBaseUrl, lobVersion, addressesRoute, lobAddressId)
	req, err := http.NewRequest("GET", getAddressUrl, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	setAuthHeaders(req, isLive)

	resp, err := l.httpClient.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// read body
	defer resp.Body.Close()
	var getAddressResponse LobGetAddressResponse
	if err := json.NewDecoder(resp.Body).Decode(&getAddressResponse); err != nil {
		return nil, err
	}
	return &getAddressResponse, nil
}

type LobDeleteAddressResponse struct {
	AddressId string `json:"id"`
	Deleted   bool   `json:"deleted"`
}

func (l *Lob) DeleteAddress(lobAddressId string, isLive bool) error {
	deleteAddressUrl := fmt.Sprintf("%s/%s/%s/%s", lobAddressBaseUrl, lobVersion, addressesRoute, lobAddressId)
	req, err := http.NewRequest("DELETE", deleteAddressUrl, nil)
	if err != nil {
		log.Println(err)
		return err
	}
	setAuthHeaders(req, isLive)

	resp, err := l.httpClient.Do(req)
	if err != nil {
		log.Println(err)
		return err
	}

	var deleteAddressResponse LobDeleteAddressResponse
	if err := json.NewDecoder(resp.Body).Decode(&deleteAddressResponse); err != nil {
		log.Println(err)
		return err
	}
	log.Println(deleteAddressResponse)

	return nil
}

type LobCreateAddressRequestMetadata struct {
	RCId string `json:"rc_id"`
}

type LobCreateAddressRequest struct {
	Name         string                          `json:"name"`
	AddressLine1 string                          `json:"address_line1"`
	AddressLine2 string                          `json:"address_line2"`
	AddressCity  string                          `json:"address_city"`
	AddressState string                          `json:"address_state"`
	AddressZip   string                          `json:"address_zip"`
	Metadata     LobCreateAddressRequestMetadata `json:"metadata"`
}

type LobCreateAddressResponse struct {
	AddressId    string `json:"id"`
	Name         string `json:"name"`
	AddressLine1 string `json:"address_line1"`
	AddressLine2 string `json:"address_line2"`
	AddressCity  string `json:"address_city"`
	AddressState string `json:"address_state"`
	AddressZip   string `json:"address_zip"`
}

func (l *Lob) CreateAddress(name, addressLine1, addressLine2, city, state, zipCode string, rcId int, isLive bool) (*LobCreateAddressResponse, error) {
	createAddressRequest := &LobCreateAddressRequest{
		Name:         name,
		AddressLine1: addressLine1,
		AddressLine2: addressLine2,
		AddressCity:  city,
		AddressState: state,
		AddressZip:   zipCode,
		Metadata:     LobCreateAddressRequestMetadata{RCId: strconv.Itoa(rcId)},
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

	setAuthHeaders(req, isLive)

	resp, err := l.httpClient.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var createAddressResponse LobCreateAddressResponse
	if err := json.NewDecoder(resp.Body).Decode(&createAddressResponse); err != nil {
		log.Println(err)
		return nil, err
	}

	return &createAddressResponse, nil
}

type LobCreatePostcardMetadata struct {
	ToRcId   string `json:"to_rc_id"`
	FromRcId string `json:"from_rc_id"`
}

type LobCreatePostcardResponse struct {
	Url string `json:"url"`
}

type LobAddress struct {
	AddressId      string `json:"id"`
	Name           string `json:"name"`
	AddressLine1   string `json:"address_line1"`
	AddressLine2   string `json:"address_line2"`
	AddressCity    string `json:"address_city"`
	AddressState   string `json:"address_state"`
	AddressZip     string `json:"address_zip"`
	AddressCountry string `json:"address_country"`
}

// https://gist.github.com/andrewmilson/19185aab2347f6ad29f5
// https://gist.github.com/mattetti/5914158/f4d1393d83ebedc682a3c8e7bdc6b49670083b84
func (l *Lob) CreatePostCard(fromLobAddress LobAddress, toLobAddress LobAddress, frontImage []byte, back string, isLive bool, fromRcId, toRcId int) (*LobCreatePostcardResponse, *LobError) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	frontPart, _ := writer.CreateFormFile("front", "user-upload")
	io.Copy(frontPart, bytes.NewReader(frontImage))

	back = fmt.Sprintf("<body>%s</body", back)

	_ = writer.WriteField("back", back)

	if fromLobAddress.AddressId != "" {
		_ = writer.WriteField("from", fromLobAddress.AddressId)

	} else {
		_ = writer.WriteField("from[name]", fromLobAddress.Name)
		_ = writer.WriteField("from[address_line1]", fromLobAddress.AddressLine1)
		_ = writer.WriteField("from[address_line2]", fromLobAddress.AddressLine2)
		_ = writer.WriteField("from[address_city]", fromLobAddress.AddressCity)
		_ = writer.WriteField("from[address_state]", fromLobAddress.AddressState)
		_ = writer.WriteField("from[address_zip]", fromLobAddress.AddressZip)
	}

	if toLobAddress.AddressId != "" {
		_ = writer.WriteField("to", toLobAddress.AddressId)
	} else {
		_ = writer.WriteField("to[name]", toLobAddress.Name)
		_ = writer.WriteField("to[address_line1]", toLobAddress.AddressLine1)
		_ = writer.WriteField("to[address_line2]", toLobAddress.AddressLine2)
		_ = writer.WriteField("to[address_city]", toLobAddress.AddressCity)
		_ = writer.WriteField("to[address_state]", toLobAddress.AddressState)
		_ = writer.WriteField("to[address_zip]", toLobAddress.AddressZip)
	}

	postcardMetadata, err := json.Marshal(LobCreatePostcardMetadata{FromRcId: strconv.Itoa(fromRcId), ToRcId: strconv.Itoa(toRcId)})
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
	setAuthHeaders(req, isLive)

	resp, err := l.httpClient.Do(req)
	if err != nil {
		return nil, &LobError{Err: err}
	}

	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		// update to read from actual response
		var createPostcardResponse LobCreatePostcardResponse
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

func setAuthHeaders(req *http.Request, isLive bool) {
	var authHeader string
	if isLive {
		authHeader = fmt.Sprintf("Basic %s",
			base64.StdEncoding.EncodeToString(
				[]byte(fmt.Sprintf("%s:", os.Getenv("LOB_API_TEST_KEY")))))
	} else {
		authHeader = fmt.Sprintf("Basic %s",
			base64.StdEncoding.EncodeToString(
				[]byte(fmt.Sprintf("%s:", os.Getenv("LOB_API_TEST_KEY")))))
	}
	req.Header.Set("Authorization", authHeader)
}
