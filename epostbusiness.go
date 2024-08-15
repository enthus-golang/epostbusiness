package epostbusiness

import "net/http"

const url = "https://api.epost.docuguide.com"

type API struct {
	Client *http.Client

	jwt string
}

func New() *API {
	return &API{
		Client: http.DefaultClient,
	}
}
