package epostbusiness

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type Letter struct {
	FileName               string `json:"fileName" validate:"min=5,max=200"`
	Data                   string `json:"data"`
	IsColor                bool   `json:"isColor"`
	IsDuplex               bool   `json:"isDuplex"`
	BatchID                int    `json:"batchID"`
	TestFlag               bool   `json:"testFlag"`
	TestShowRestrictedArea bool   `json:"testShowRestrictedArea"`
	TestEMail              string `json:"testEMail"`
	AddressLine1           string `json:"addressLine1"`
	AddressLine2           string `json:"addressLine2"`
	AddressLine3           string `json:"addressLine3"`
	ZipCode                string `json:"zipCode"`
	City                   string `json:"city"`
	Country                string `json:"country"`
	SenderAdressLine1      string `json:"senderAdressLine1"`
	SenderStreet           string `json:"senderStreet"`
	SenderZipCode          string `json:"senderZipCode"`
	SenderCity             string `json:"senderCity"`
	Custom1                string `json:"custom1"`
	Custom2                string `json:"custom2"`
}

type LetterIdentifier struct {
	FileName string `json:"fileName"`
	LetterID int    `json:"letterID"`
}

func (a API) CreateLetters(ctx context.Context, letters []Letter) ([]LetterIdentifier, error) {
	var letterIdentifiers []LetterIdentifier

	body, err := json.Marshal(letters)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url+"/api/Letter", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.jwt)

	res, err := a.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusOK:
		err = json.NewDecoder(res.Body).Decode(&letterIdentifiers)
		if err != nil {
			return nil, err
		}

		return letterIdentifiers, nil

	case http.StatusBadRequest:
		fallthrough
	case http.StatusUnauthorized:
		fallthrough
	case http.StatusTooManyRequests:
		var loginErr loginError

		err = json.NewDecoder(res.Body).Decode(&loginErr)
		if err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("%s: %s", loginErr.Code, loginErr.Description)

	default:
		return nil, errors.New(res.Status)
	}
}

type LetterStatus struct {
	LetterID                   int           `json:"letterID"`
	FileName                   string        `json:"fileName"`
	StatusID                   int           `json:"statusID"`
	StatusDetails              *string       `json:"statusDetails"`
	CreatedDate                Time          `json:"createdDate"`
	ProcessedDate              *Time         `json:"processedDate"`
	PrintUploadDate            *Time         `json:"printUploadDate"`
	PrintFeedbackDate          *Time         `json:"printFeedbackDate"`
	TestFlag                   *bool         `json:"testFlag"`
	TestEMail                  *string       `json:"testEMail"`
	TestShowRestrictedArea     bool          `json:"testShowRestrictedArea"`
	RegisteredLetter           *string       `json:"registeredLetter"`
	RegisteredLetterID         *string       `json:"registeredLetterID"`
	BatchID                    int           `json:"batchID"`
	CoverLetter                bool          `json:"coverLetter"`
	NoOfPages                  int           `json:"noOfPages"`
	SubVendorID                *string       `json:"subVendorID"`
	Custom1                    *string       `json:"custom1"`
	Custom2                    *string       `json:"custom2"`
	Custom3                    *string       `json:"custom3"`
	Custom4                    *string       `json:"custom4"`
	Custom5                    *string       `json:"custom5"`
	ZipCode                    *string       `json:"zipCode"`
	City                       *string       `json:"city"`
	Country                    *string       `json:"country"`
	IsColor                    bool          `json:"isColor"`
	IsDuplex                   bool          `json:"isDuplex"`
	RegisteredLetterStatus     *string       `json:"registeredLetterStatus"`
	RegisteredLetterStatusDate *Time         `json:"registeredLetterStatusDate"`
	ErrorList                  []LetterError `json:"errorList"`
}

type LetterError struct {
	Description string `json:"description"`
	Level       string `json:"level"`
	Code        string `json:"code"`
	Date        *Time  `json:"date"`
}

func (a API) GetLettersStatusList(ctx context.Context, letterIDs []int) ([]LetterStatus, error) {
	var statusList []LetterStatus

	body, err := json.Marshal(letterIDs)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url+"/api/Letter/StatusQuery", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.jwt)

	res, err := a.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%d: %s", res.StatusCode, res.Status)
	}

	err = json.NewDecoder(res.Body).Decode(&statusList)
	if err != nil {
		return nil, err
	}

	return statusList, nil
}
