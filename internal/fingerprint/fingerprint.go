package fingerprint

import (
	"context"
	"encoding/json"
	"fmt"
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

func (c *Client) Check(requestId, visitorId string) (passed bool, err error) {

	// Configure authorization, in our case with API Key
	auth := context.WithValue(context.Background(), sdk.ContextAPIKey, sdk.APIKey{
		Key: c.APIKey,
	})

	log.Printf("Checking request %s with API key %s in region %s\n", requestId, c.APIKey, c.Cfg.GetRegion())

	response, httpRes, err := c.API.FingerprintApi.GetEvent(auth, requestId)

	// See all the data that you can run verifications against
	r, _ := json.MarshalIndent(response, "", "\t")
	log.Printf("%v\n", string(r))

	if err != nil || httpRes.StatusCode != 200 {
		return false, fmt.Errorf("FingerprintApi.GetEvent: HTTP %d: %w\n", httpRes.StatusCode, err)
	}

	// Compare the fingerprints, to detect if the fingerprint received from the browser has been tampered with

	if response.Products.Identification.Data.VisitorId != visitorId {
		return false, fmt.Errorf("fingerprint mismatch: expected %s, got %s\n", visitorId, response.Products.Identification.Data.VisitorId)
	}

	return true, nil
}
