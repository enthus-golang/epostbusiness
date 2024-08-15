package epostbusiness

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type login struct {
	VendorID string `json:"vendorID"`
	EKP      string `json:"ekp"`
	Secret   string `json:"secret"`
	Password string `json:"password"`
}

type loginToken struct {
	Token string `json:"token"`
}

type loginError struct {
	Level       string    `json:"level"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
}

func (a *API) Login(ctx context.Context, vendorID, ekp, secret, password string) (bool, error) {
	body, err := json.Marshal(login{
		VendorID: vendorID,
		EKP:      ekp,
		Secret:   secret,
		Password: password,
	})
	if err != nil {
		return false, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url+"/api/Login", bytes.NewReader(body))
	if err != nil {
		return false, err
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := a.Client.Do(req)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusOK:
		var token loginToken
		err = json.NewDecoder(res.Body).Decode(&token)
		if err != nil {
			return false, err
		}

		a.jwt = token.Token
		return true, nil

	case http.StatusBadRequest:
		fallthrough
	case http.StatusUnauthorized:
		fallthrough
	case http.StatusTooManyRequests:
		var loginErr loginError
		err = json.NewDecoder(res.Body).Decode(&loginErr)
		if err != nil {
			return false, err
		}

		return false, fmt.Errorf("%s: %s", loginErr.Code, loginErr.Description)

	default:
		return false, errors.New(res.Status)
	}
}
