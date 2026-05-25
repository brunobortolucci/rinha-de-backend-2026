package vector

import "time"

func clamp(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}

func Vectorize(req Request, mccRisk map[string]float64, norm Normalization) ([14]float64, error) {
	var v [14]float64
	v[0] = clamp(req.Transaction.Amount / norm.MaxAmount)
	v[1] = clamp(float64(req.Transaction.Installments) / norm.MaxInstallments)
	v[2] = clamp((req.Transaction.Amount / req.Customer.AvgAmount) / norm.AmountVsAvgRatio)

	t, err := time.Parse(time.RFC3339, req.Transaction.RequestedAt)
	if err != nil {
		t = time.Now()
	}
	v[3] = float64(t.Hour()) / 23

	week := int(t.Weekday())
	dayOfWeek := (week + 6) % 7
	v[4] = float64(dayOfWeek) / 6

	if req.LastTransaction == nil {
		v[5] = -1
		v[6] = -1
	} else {
		lastTs, err := time.Parse(time.RFC3339, req.LastTransaction.Timestamp)
		if err != nil {
			lastTs = time.Now()
		}
		minutes := t.Sub(lastTs).Minutes()
		v[5] = clamp(minutes / norm.MaxMinutes)
		v[6] = clamp(req.LastTransaction.KmFromCurrent / norm.MaxKm)
	}

	v[7] = clamp(req.Terminal.KmFromHome / norm.MaxKm)
	v[8] = clamp(float64(req.Customer.TxCount24h) / norm.MaxTxCount24h)

	if req.Terminal.IsOnline {
		v[9] = 1
	}

	if req.Terminal.CardPresent {
		v[10] = 1
	}

	known := false
	for _, m := range req.Customer.KnownMerchants {
		if m == req.Merchant.ID {
			known = true
			break
		}
	}
	if !known {
		v[11] = 1
	}

	risk, ok := mccRisk[req.Merchant.MCC]
	if !ok {
		risk = 0.5
	}
	v[12] = risk

	v[13] = clamp(req.Merchant.AvgAmount / norm.MaxMerchantAvgAmount)

	return v, nil
}
