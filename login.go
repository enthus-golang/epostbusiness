package epostbusiness

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
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
	defer func() {
		if cerr := res.Body.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	if res.StatusCode != http.StatusOK {
		return false, newAPIError(res)
	}

	var token loginToken
	if err = json.NewDecoder(res.Body).Decode(&token); err != nil {
		return false, err
	}

	a.jwt = token.Token

	return true, nil
}
