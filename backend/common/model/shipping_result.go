package model

// ShippingResult 物流费率诊断结果
type ShippingResult struct {
	RecommendedCode string         `json:"recommended_code"`
	Rates           []ShippingRate `json:"rates"`
}

// ShippingRate 单个物流费率
type ShippingRate struct {
	Carrier     string   `json:"carrier"`
	Service     string   `json:"service"`
	TotalFee    float64  `json:"total_fee"`
	TransitDays int      `json:"transit_days"`
	Tags        []string `json:"tags"` // CHEAPEST/FASTEST
}
