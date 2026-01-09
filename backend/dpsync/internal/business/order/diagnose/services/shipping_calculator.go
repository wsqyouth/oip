package services

import (
	"context"
	"fmt"
	"hash/fnv"
	"math/rand"

	"oip/common/model"
)

// ShippingCalculator 物流费率计算器（Mock）
type ShippingCalculator struct{}

// NewShippingCalculator 创建费率计算器实例
func NewShippingCalculator() *ShippingCalculator {
	return &ShippingCalculator{}
}

// Calculate 计算物流费率（Mock - 基于 order_id 生成确定性伪随机费率）
// shipment 参数包含物流信息,未来可用于更精确的费率计算
func (c *ShippingCalculator) Calculate(ctx context.Context, orderID string, shipment map[string]interface{}) (*model.ShippingResult, error) {
	// 1. 基于 order_id 生成确定性种子
	seed := hashSeed(orderID)
	rng := rand.New(rand.NewSource(seed))

	// 2. Mock 承运商费率表
	carriers := []struct {
		Carrier     string
		Service     string
		BaseRate    float64
		TransitDays int
	}{
		{"FedEx", "Ground", 12.50, 3},
		{"UPS", "Ground", 15.20, 3},
		{"USPS", "Priority", 9.80, 5},
		{"DHL", "Express", 28.00, 1},
	}

	// 3. 生成费率（加入随机波动，但同一 order_id 结果一致）
	rates := make([]model.ShippingRate, 0, len(carriers))
	for _, carrier := range carriers {
		// 随机波动 ±10%
		fluctuation := carrier.BaseRate * (rng.Float64() - 0.5) * 0.2
		totalFee := carrier.BaseRate + fluctuation

		rates = append(rates, model.ShippingRate{
			Carrier:     carrier.Carrier,
			Service:     carrier.Service,
			TotalFee:    roundTo2Decimals(totalFee),
			TransitDays: carrier.TransitDays,
			Tags:        []string{},
		})
	}

	// 4. 标记 CHEAPEST 和 FASTEST
	cheapestIdx := findCheapest(rates)
	fastestIdx := findFastest(rates)
	rates[cheapestIdx].Tags = append(rates[cheapestIdx].Tags, "CHEAPEST")
	rates[fastestIdx].Tags = append(rates[fastestIdx].Tags, "FASTEST")

	// 5. 推荐最便宜的
	recommendedCode := fmt.Sprintf("%s_%s", rates[cheapestIdx].Carrier, rates[cheapestIdx].Service)

	return &model.ShippingResult{
		RecommendedCode: recommendedCode,
		Rates:           rates,
	}, nil
}

// hashSeed 基于字符串生成确定性种子
func hashSeed(s string) int64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return int64(h.Sum64())
}

// roundTo2Decimals 四舍五入到两位小数
func roundTo2Decimals(f float64) float64 {
	return float64(int(f*100+0.5)) / 100
}

// findCheapest 找到最便宜的费率索引
func findCheapest(rates []model.ShippingRate) int {
	minIdx := 0
	minFee := rates[0].TotalFee
	for i, rate := range rates {
		if rate.TotalFee < minFee {
			minFee = rate.TotalFee
			minIdx = i
		}
	}
	return minIdx
}

// findFastest 找到最快的费率索引
func findFastest(rates []model.ShippingRate) int {
	minIdx := 0
	minDays := rates[0].TransitDays
	for i, rate := range rates {
		if rate.TransitDays < minDays {
			minDays = rate.TransitDays
			minIdx = i
		}
	}
	return minIdx
}
