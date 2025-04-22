package entity

import (
	"time"

	"github.com/LucianoGiope/posgolangdesafio1/pkg/entity"
)

type DollarQuote struct {
	ID        string `json:"id"`
	Value     string `json:"value"`
	CreatedAt string `json:"created_at"`
}

func NewDollarQuote(value string) (*DollarQuote, error) {
	dateQuoteNow := time.Now()
	dollarQuote := &DollarQuote{
		ID:        entity.NewUUID(),
		Value:     value,
		CreatedAt: dateQuoteNow.Format("02/01/2006 15:04:05"),
	}
	return dollarQuote, nil
}
