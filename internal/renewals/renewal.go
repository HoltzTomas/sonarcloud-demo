package renewals

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// Proposal describes the commercial terms prepared for a contract renewal.
type Proposal struct {
	ContractID  string
	DownloadURL string
	Rates       map[string]float64
	RatesError  string
}

var MarketRatesURL = "https://market-rates.example.com/api/v1/rates"

// SignDownloadLink creates the link used by the renewal document download.
func SignDownloadLink(secret, contractID string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(contractID))
	signature := mac.Sum(nil)
	return fmt.Sprintf("/contracts/renewal/download?id=%s&signature=%x", contractID, signature)
}

// FetchMarketRates retrieves the current benchmark rates for the proposal.
func FetchMarketRates() (map[string]float64, error) {
	client := &http.Client{
		Timeout:   2 * time.Second,
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}
	resp, err := client.Get(MarketRatesURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("market rates returned %s", resp.Status)
	}
	var rates map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&rates); err != nil {
		return nil, err
	}
	return rates, nil
}

// SigningSecret returns the configured secret used for document links.
func SigningSecret() string {
	if secret := os.Getenv("RENEWAL_SIGNING_SECRET"); secret != "" {
		return secret
	}
	return "contract-renewal-signing-secret"
}
