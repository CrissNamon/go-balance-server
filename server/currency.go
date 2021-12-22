package server

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	CURRENCY_RATES_API           string = "https://api.exchangerate.host/latest"
	CURRENCY_RATES_API_CONVERTER string = CURRENCY_RATES_API + "?base=%s&symbols=%s"
	BASE_CURRENCY                string = "RUB"
)

func GetCurrencyRate(base string, to string) (float64, error) {
	client := http.Client{}
	apiUrl := fmt.Sprintf(CURRENCY_RATES_API_CONVERTER, base, to)
	request, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return 0, err
	}

	resp, err := client.Do(request)
	if err != nil {
		return 0, err
	}

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return 0, err
	}
	rates, ok := result["rates"].(map[string]interface{})
	if !ok {
		return 0, &OperationError{ERROR_INTERNAL}
	}
	rate, ok := rates[to].(float64)
	if !ok {
		return 0, &OperationError{ERROR_BALANCE_WRONG_CURRENCY_CODE}
	}
	return rate, nil
}
