package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
	Url            string `json:"url"`
	Name           string `json:"name"`
	AddressLine1   string `json:"address_line1"`
	AddressLine2   string `json:"address_line2"`
	AddressCity    string `json:"address_city"`
	AddressState   string `json:"address_state"`
	AddressZip     string `json:"address_zip"`
	AddressCountry string `json:"address_country"`
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
		panic(err)
	}

	// read body
	defer resp.Body.Close()
	var getAddressResponse GetAddressResponse
	if err := json.NewDecoder(resp.Body).Decode(&getAddressResponse); err != nil {
		return nil, err
	}
	return &getAddressResponse, nil
}
