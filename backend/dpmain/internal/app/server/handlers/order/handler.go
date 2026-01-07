package order

import "oip/dpmain/internal/app/domains/services/svorder"

// OrderHandler 订单 HTTP 处理器
type OrderHandler struct {
	orderService *svorder.OrderService
}

// NewOrderHandler 创建订单处理器实例
func NewOrderHandler(orderService *svorder.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}
