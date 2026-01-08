package order

import (
	"log"

	"github.com/gin-gonic/gin"
	"oip/dpmain/internal/app/domains/apimodel/response"
	"oip/dpmain/internal/app/pkg/ginx"
)

// Get godoc
// @Summary      获取订单详情
// @Description  根据订单ID获取订单详细信息（包含诊断结果）
// @Description
// @Description  使用场景：
// @Description  - 创建订单返回 code=3001 时，通过此接口轮询结果
// @Description  - 查询历史订单详情
// @Tags         orders
// @Produce      json
// @Param        id path string true "订单ID（UUID）"
// @Success      200 {object} ginx.Response{data=response.OrderResponse} "查询成功"
// @Failure      400 {object} ginx.Response "参数错误"
// @Failure      404 {object} ginx.Response "订单不存在"
// @Failure      500 {object} ginx.Response "服务器错误"
// @Security     ApiKeyAuth
// @Router       /orders/{id} [get]
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
