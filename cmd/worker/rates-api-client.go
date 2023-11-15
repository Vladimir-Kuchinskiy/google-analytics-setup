package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"
)

type RatesAPIClient struct {
	client http.Client
}

func NewRatesAPIClient() *RatesAPIClient {
	return &RatesAPIClient{
		client: http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type CurrenciesRatio struct {
	LeftCurrency  string
	RightCurrency string
	Rate          float64
	Date          string
}

func (c *RatesAPIClient) fetchUAHtoUSDCurrenciesRatio() (CurrenciesRatio, error) {
	request, err := http.NewRequest(http.MethodGet, "https://bank.gov.ua/NBUStatService/v1/statdirectory/exchange?json", nil)
	if err != nil {
		return CurrenciesRatio{}, err
	}

	response, err := c.client.Do(request)
	if err != nil {
		return CurrenciesRatio{}, err
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return CurrenciesRatio{}, err
	}

	var result []struct {
		Currency string  `json:"cc"`
		Date     string  `json:"exchangedate"`
		Rate     float64 `json:"rate"`
	}
	if err = json.Unmarshal(responseBody, &result); err != nil {
		return CurrenciesRatio{}, err
	}

	for _, el := range result {
		if el.Currency == "USD" {
			return CurrenciesRatio{
				LeftCurrency:  "UAH",
				RightCurrency: "USD",
				Rate:          el.Rate,
				Date:          el.Date,
			}, nil
		}
	}

	return CurrenciesRatio{}, errors.New("currencies ratio for USD/UAH not found")

}
