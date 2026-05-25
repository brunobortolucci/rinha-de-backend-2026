package config

import (
	"encoding/json"
	"fmt"

	"github.com/brunobortolucci/rinha-de-backend-2026/internal/vector"
)

type MccRisk map[string]float64

func LoadNormalization(data []byte) (vector.Normalization, error) {
	var norm vector.Normalization
	if err := json.Unmarshal(data, &norm); err != nil {
		return vector.Normalization{}, fmt.Errorf("unmarshal normalização: %w", err)
	}

	return norm, nil
}

func LoadMcc(data []byte) (MccRisk, error) {
	var mcc MccRisk
	if err := json.Unmarshal(data, &mcc); err != nil {
		return nil, fmt.Errorf("unmarshal mmc risk: %w", err)
	}

	return mcc, nil
}
