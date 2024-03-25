package fingerprint

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/fingerprintjs/fingerprint-pro-server-api-go-sdk/v5/sdk"
)

type Client struct {
	API    *sdk.APIClient
	Cfg    *sdk.Configuration
	APIKey string
}

func New() *Client {
	cfg := sdk.NewConfiguration()
	client := sdk.NewAPIClient(cfg)

	// Default region is sdk.RegionUS
	if strings.ToLower(os.Getenv("FINGERPRINT_REGION")) == "eu" {
		cfg.ChangeRegion(sdk.RegionEU)
	}
	if strings.ToLower(os.Getenv("FINGERPRINT_REGION")) == "ap" {
		cfg.ChangeRegion(sdk.RegionAsia)
	}

	return &Client{
		API:    client,
		Cfg:    cfg,
		APIKey: os.Getenv("FINGERPRINT_SECRET_KEY"),
	}
}

func (c *Client) Check(requestId string) string {

	// Configure authorization, in our case with API Key
	auth := context.WithValue(context.Background(), sdk.ContextAPIKey, sdk.APIKey{
		Key: c.APIKey,
	})

	log.Printf("Checking request %s with API key %s in region %s\n", requestId, c.APIKey, c.Cfg.GetRegion())

	response, httpRes, err := c.API.FingerprintApi.GetEvent(auth, requestId)

	log.Printf("HTTP response: %+v\n", httpRes)

	if err != nil {
		log.Fatalf("FingerprintApi.GetEvent: %s\n", err)
	}

	if response.Products.Botd != nil {
		log.Printf("Got response with Botd: %v \n", response.Products.Botd)
	}

	if response.Products.Identification == nil {
		return ""
	}
	stringResponse, _ := json.Marshal(response.Products.Identification)
	return string(stringResponse)

}
