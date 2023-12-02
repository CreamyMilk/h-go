package notifiers

import (
	"encoding/json"
	"fmt"
	"io"

	b64 "encoding/base64"
	"github.com/rs/zerolog/log"
	"hosigo/utils"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type SmsType string

const (
	OTP            SmsType = "OTP"
	RecipientAlert SmsType = "RecipientAlert"
)

func toBasicBase64AuthValue(username, password string) string {
	message := username + ":" + password
	encodedStr := b64.StdEncoding.EncodeToString([]byte(message))
	return "Basic " + encodedStr
}

// Usage requires the following envs:
//
//	TWILIO_BASE_URL, TWILIO_PHONE_NUMBER (optional)
//	TWILIO_ACCOUNT_SID, TWILIO_AUTH_TOKEN (required)

func SendSms(message, phoneNumber string, smsType SmsType) error {
	if !utils.CheckEnvInclusion("TWILIO_ACCOUNT_SID", "TWILIO_NUMBER_SID", "TWILIO_AUTH_TOKEN", "TWILIO_PHONE_NUMBER") {
		log.Printf("Issue with Twilio Envs was to send %s to (%s) \n", message, phoneNumber)
		return fmt.Errorf("invalid OTP request. Contact Abacus support")
	}

	baseUrl := utils.UseEnvOrDefault("TWILIO_BASE_URL", "https://api.twilio.com/2010-04-01")
	endpoint := fmt.Sprintf("%s/Accounts/%s/Messages.json", baseUrl, os.Getenv("TWILIO_ACCOUNT_SID"))

	senderNumber := utils.UseEnvOrDefault("TWILIO_PHONE_NUMBER", "+18775657133")

	data := url.Values{}
	data.Set("Body", message)
	data.Set("To", phoneNumber)
	data.Set("From", senderNumber)

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", toBasicBase64AuthValue(os.Getenv("TWILIO_NUMBER_SID"), os.Getenv("TWILIO_AUTH_TOKEN")))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var parsedResponse twilioSmsResponseWrapper
	err = json.Unmarshal(responseBody, &parsedResponse)
	if err != nil {
		return err
	}

	if parsedResponse.RequestErrorCode != nil || parsedResponse.RequestErrorMessage != nil {
		if parsedResponse.RequestErrorCode != nil {
			return fmt.Errorf("invalid request made to the messaging service: %v", *parsedResponse.RequestErrorMessage)
		}
		return fmt.Errorf("invalid request made to the messaging service")
	}

	if parsedResponse.ErrorCode != nil || parsedResponse.ErrorMessage != nil {
		if parsedResponse.ErrorMessage != nil {
			return fmt.Errorf("unexpected response received from messaging service: %v", *parsedResponse.ErrorMessage)
		}
		return fmt.Errorf("unexpected response received from messaging service")
	}
	return nil
}

type twilioSmsResponseWrapper struct {
	RequestErrorCode    *int64      `json:"code"`
	RequestErrorMessage *string     `json:"message"`
	AccountSid          string      `json:"account_sid"`
	APIVersion          string      `json:"api_version"`
	Body                string      `json:"body"`
	DateCreated         string      `json:"date_created"`
	DateSent            string      `json:"date_sent"`
	DateUpdated         string      `json:"date_updated"`
	Direction           string      `json:"direction"`
	ErrorCode           *string     `json:"error_code"`
	ErrorMessage        *string     `json:"error_message"`
	From                string      `json:"from"`
	MessagingServiceSid *string     `json:"messaging_service_sid"`
	NumMedia            string      `json:"num_media"`
	NumSegments         string      `json:"num_segments"`
	Price               *string     `json:"price"`
	PriceUnit           string      `json:"price_unit"`
	Sid                 string      `json:"sid"`
	Status              interface{} `json:"status"`
	SubresourceUris     struct {
		Media string `json:"media"`
	} `json:"subresource_uris"`
	To  string `json:"to"`
	URI string `json:"uri"`
}
