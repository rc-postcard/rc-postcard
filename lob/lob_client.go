package lob

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
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
	RecurseCenterName     = "Recurse Center"
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
const verificationsRoute = "us_verifications"

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
			Mode     string `json:"mode"`
		} `json:"metadata"`
		DateCreated          time.Time `json:"date_created"`
		ExpectedDeliveryDate string    `json:"expected_delivery_date"`
	} `json:"data"`
}

func (l *Lob) GetPostcards(recipientRecurseId int, isLive bool) (*LobGetPostcardsResponse, error) {
	getPostcardsUrl := fmt.Sprintf("%s/%s/%s?metadata[to_rc_id]=%d", lobAddressBaseUrl, lobVersion, postcardsRoute, recipientRecurseId)
	req, err := http.NewRequest("GET", getPostcardsUrl, nil)
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
func (l *Lob) CreatePostCard(fromLobAddress LobAddress, toLobAddress LobAddress, frontImage []byte, back string, isLive bool, fromRcId, toRcId int, mode string) (*LobCreatePostcardResponse, *LobError) {
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

	_ = writer.WriteField("metadata[to_rc_id]", strconv.Itoa(toRcId))
	_ = writer.WriteField("metadata[from_rc_id]", strconv.Itoa(fromRcId))
	_ = writer.WriteField("metadata[mode]", mode)

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

type LobVerifyAddressRequest struct {
	PrimaryLine   string `json:"primary_line"`
	SecondaryLine string `json:"secondary_line"`
	City          string `json:"city"`
	State         string `json:"state"`
	ZipCode       string `json:"zip_code"`
}

type LobVerifyAddressResponse struct {
	Deliverability string `json:"deliverability"`
}

const (
	Deliverable                = "deliverable"
	DeliverableUnnecessaryUnit = "deliverable_unnecessary_unit"
	DeliverableIncorrectUnit   = "deliverable_incorrect_unit"
	DeliverableMissingUnit     = "deliverable_missing_unit"
	Undeliverable              = "undeliverable"
)

func (l *Lob) VerifyAddress(addressLine1, addressLine2, city, state, zipCode string) (*LobVerifyAddressResponse, error) {
	verifyAddressRequest := &LobVerifyAddressRequest{
		PrimaryLine:   addressLine1,
		SecondaryLine: addressLine2,
		City:          city,
		State:         state,
		ZipCode:       zipCode,
	}

	log.Printf("PRIMARY %s SECONDARY %s CITY %s STATE %s ZIPCODE %s\n",
		verifyAddressRequest.PrimaryLine,
		verifyAddressRequest.SecondaryLine,
		verifyAddressRequest.City,
		verifyAddressRequest.State,
		verifyAddressRequest.ZipCode)

	marshalledVerifyAddressRequest, err := json.Marshal(verifyAddressRequest)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	verifyAddressUrl := fmt.Sprintf("%s/%s/%s", lobAddressBaseUrl, lobVersion, verificationsRoute)
	req, err := http.NewRequest("POST", verifyAddressUrl, bytes.NewBuffer(marshalledVerifyAddressRequest))
	if err != nil {
		log.Println(err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	setAuthHeaders(req, true)

	resp, err := l.httpClient.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var verifyAddressResponse LobVerifyAddressResponse
	if err := json.NewDecoder(resp.Body).Decode(&verifyAddressResponse); err != nil {
		log.Println(err)
		return nil, err
	}

	return &verifyAddressResponse, nil
}

func (l *Lob) VerifyAddressBySendingTestPostcard(addressLine1, addressLine2, city, state, zipCode string) (*LobVerifyAddressResponse, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	_ = writer.WriteField("front", fmt.Sprintf("<body>Hello!</body>"))
	_ = writer.WriteField("back", fmt.Sprintf("<body>Goodbye!</body>"))

	_ = writer.WriteField("from[name]", "Recurse Center")
	_ = writer.WriteField("from[address_line1]", RecurseAddressLine1)
	_ = writer.WriteField("from[address_line2]", RecurseAddressLine1)
	_ = writer.WriteField("from[address_city]", RecurseAddressCity)
	_ = writer.WriteField("from[address_state]", RecurseAddressState)
	_ = writer.WriteField("from[address_zip]", RecurseAddressZip)

	_ = writer.WriteField("to[name]", "Your name")
	_ = writer.WriteField("to[address_line1]", addressLine1)
	_ = writer.WriteField("to[address_line2]", addressLine2)
	_ = writer.WriteField("to[address_city]", city)
	_ = writer.WriteField("to[address_state]", state)
	_ = writer.WriteField("to[address_zip]", zipCode)

	_ = writer.WriteField("metadata[mode]", "ver")

	writer.Close()

	postPostcardUrl := fmt.Sprintf("%s/%s/%s", lobAddressBaseUrl, lobVersion, postcardsRoute)
	req, err := http.NewRequest("POST", postPostcardUrl, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())
	setAuthHeaders(req, false)

	resp, err := l.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		// update to read from actual response
		var createPostcardResponse LobCreatePostcardResponse
		if err := json.NewDecoder(resp.Body).Decode(&createPostcardResponse); err != nil {
			return nil, err
		}
		return &LobVerifyAddressResponse{Deliverability: Deliverable}, nil
	} else if resp.StatusCode == 422 {
		return &LobVerifyAddressResponse{Deliverability: Undeliverable}, nil
	} else {
		var lobErrorResponse LobErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&lobErrorResponse); err != nil {
			return nil, err
		}

		log.Printf("STATUS CODE %d, Response %v\n", resp.StatusCode, lobErrorResponse)
		return nil, errors.New("err_verifying_address")
	}
}

func setAuthHeaders(req *http.Request, isLive bool) {
	var authHeader string
	if isLive {
		authHeader = fmt.Sprintf("Basic %s",
			base64.StdEncoding.EncodeToString(
				[]byte(fmt.Sprintf("%s:", os.Getenv("LOB_API_LIVE_KEY")))))
	} else {
		authHeader = fmt.Sprintf("Basic %s",
			base64.StdEncoding.EncodeToString(
				[]byte(fmt.Sprintf("%s:", os.Getenv("LOB_API_TEST_KEY")))))
	}
	req.Header.Set("Authorization", authHeader)
}
