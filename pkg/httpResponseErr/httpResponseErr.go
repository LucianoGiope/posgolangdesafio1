package httpResponseErr

import (
	"encoding/json"
	"errors"
)

type QuoteError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func NewQuoteError(msg string, code int) *QuoteError {
	return &QuoteError{msg, code}
}

func (qe *QuoteError) DisplayMessage(jsonBody []byte) (string, error) {
	var dataResult = qe
	err := json.Unmarshal(jsonBody, &dataResult)
	if err != nil {
		return dataResult.Message, errors.New(err.Error())
	}
	return dataResult.Message, nil
}
