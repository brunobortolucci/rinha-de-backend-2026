package vector

type Request struct {
	Id              string           `json:"id,omitempty"`
	Transaction     Transaction      `json:"transaction,omitempty"`
	Customer        Customer         `json:"customer,omitempty"`
	Merchant        Merchant         `json:"merchant,omitempty"`
	Terminal        Terminal         `json:"terminal,omitempty"`
	LastTransaction *LastTransaction `json:"last_transaction"`
}

type Transaction struct {
	Amount       float64 `json:"amount"`
	Installments int     `json:"installments"`
	RequestedAt  string  `json:"requested_at"`
}

type Customer struct {
	AvgAmount      float64  `json:"avg_amount"`
	TxCount24h     int      `json:"tx_count_24h"`
	KnownMerchants []string `json:"known_merchants"`
}

type Merchant struct {
	ID        string  `json:"id,omitempty"`
	MCC       string  `json:"mcc,omitempty"`
	AvgAmount float64 `json:"avg_amount"`
}

type Terminal struct {
	IsOnline    bool    `json:"is_online"`
	CardPresent bool    `json:"card_present"`
	KmFromHome  float64 `json:"km_from_home"`
}

type LastTransaction struct {
	Timestamp     string  `json:"timestamp"`
	KmFromCurrent float64 `json:"km_from_current"`
}

type Response struct {
	Approved   bool    `json:"approved"`
	FraudScore float64 `json:"fraud_score"`
}
