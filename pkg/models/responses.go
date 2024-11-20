package models


type PriceResponse struct {
    Symbol  string  `json:"symbol"`
    Source  string  `json:"source"`
    Price   float64 `json:"price"`
    Elapsed float64 `json:"elapsed"`
    Note    string  `json:"note,omitempty"`
}


type Transaction struct {
	Price  interface{} `bson:"price"`
	Amount interface{} `bson:"amount"`
}

type MarketSourceResult struct {
	MarketName   string
	Source       string
	Median       float64
	WeightedMean float64
	SumAmounts   float64
}
