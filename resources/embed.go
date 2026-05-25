package resources

import _ "embed"

//go:embed normalization.json
var Normalization []byte

//go:embed mcc_risk.json
var MccRisk []byte
