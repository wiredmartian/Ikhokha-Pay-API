package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/joho/godotenv"
)

type Envs struct {
	apiUrl string
	appId  string
	secret string
}

type Urls struct {
	CallBackUrl    string `json:"callbackUrl"`
	SuccessPageUrl string `json:"successPageUrl"`
	FailurePageUrl string `json:"failurePageUrl"`
	CancelUrl      string `json:"cancelUrl"`
}
type PaymentRequest struct {
	EntityID              string `json:"entityID"`
	ExternalEntityID      string `json:"externalEntityID"`
	Amount                int64  `json:"amount"`
	Currency              string `json:"currency"`
	RequesterUrl          string `json:"requesterUrl"`
	Mode                  string `json:"mode"`
	ExternalTransactionID string `json:"externalTransactionID"`
	Urls                  Urls   `json:"urls"`
}

type PaymentResponse struct {
	ResponseCode          string `json:"responseCode"`
	Message               string `json:"message"`
	PaylinkUrl            string `json:"paylinkUrl"`
	PaylinkID             string `json:"paylinkID"`
	ExternalTransactionID string `json:"externalTransactionID"`
}

func main() {
	requestBody := PaymentRequest{
		EntityID:              "MyEntityID1234",
		ExternalEntityID:      "MARTIANS1234",
		Amount:                10000,
		Currency:              "ZAR",
		RequesterUrl:          "https://example.com/requester",
		Mode:                  "live",
		ExternalTransactionID: "TRANS789111",
		Urls: Urls{
			CallBackUrl:    "https://example.com/callback",
			SuccessPageUrl: "https://example.com/success",
			FailurePageUrl: "https://example.com/failure",
			CancelUrl:      "https://example.com/cancel",
		},
	}

	fmt.Println("Creating paylink...")

	response, err := CreatePayLink(requestBody)
	if err != nil {
		fmt.Println("Error creating paylink: ", err)
		return
	}
	fmt.Println("Paylink created successfully: ", response.PaylinkUrl)
}

func CreatePayLink(request PaymentRequest) (*PaymentResponse, error) {
	envs := getEnvs()
	fmt.Println("Creating paylink with request:", request)

	payloadBytes, err := json.Marshal(request)
	if err != nil {
		fmt.Println("Error marshalling request body:", err)
		return nil, err
	}

	req, err := http.NewRequest("POST", envs.apiUrl, bytes.NewBuffer(payloadBytes))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil, err
	}

	signature, err := generateSignature(string(payloadBytes), envs.apiUrl, envs.secret)
	if err != nil {
		fmt.Println("Error generating signature:", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("IK-APPID", envs.appId)
	req.Header.Set("IK-SIGN", signature)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return nil, err
	}
	defer resp.Body.Close()

	var response PaymentResponse
	json.NewDecoder(resp.Body).Decode(&response)
	return &response, nil
}

// Get all needed environment Variables
func getEnvs() Envs {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file", err)
	}
	apiUrl := os.Getenv("IK_API_URL")
	appId := os.Getenv("IK_APP_ID")
	secret := os.Getenv("IK_APP_SECRET")

	if apiUrl == "" || appId == "" || secret == "" {
		fmt.Println("Missing required environment variables")
		os.Exit(1)
	}
	return Envs{apiUrl, appId, secret}
}

// Generate the signature for the request
func generateSignature(body string, endpoint string, secret string) (string, error) {
	parsedUrl, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}

	basePath := parsedUrl.Path
	if basePath == "" {
		return "", fmt.Errorf("no basePath in url")
	}

	sanitizedBody := jsonEscape(basePath + body)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(sanitizedBody))
	return hex.EncodeToString(mac.Sum(nil)), nil
}

// Escape the JSON string
func jsonEscape(payload string) string {
	var escaped string
	for _, char := range payload {
		switch char {
		case '\\', '"', '\'':
			escaped += "\\" + string(char)
		case '\u0000':
			escaped += "\\0"
		default:
			escaped += string(char)
		}
	}
	return escaped
}
