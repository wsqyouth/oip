package order

import (
	"log"

	"github.com/gin-gonic/gin"
	"oip/dpmain/internal/app/domains/apimodel/response"
	"oip/dpmain/internal/app/pkg/ginx"
)

// Get 查询订单接口
// GET /api/v1/orders/:id
func (h *OrderHandler) Get(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		ginx.BadRequest(c, "order_id required")
		return
	}

	order, err := h.orderService.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		log.Printf("[ERROR] get order failed: %v", err)
		ginx.NotFound(c, "order not found")
		return
	}

	ginx.Success(c, response.FromOrderEntity(order))
}
