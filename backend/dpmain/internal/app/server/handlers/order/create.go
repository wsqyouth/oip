package order

import (
	"fmt"
	"log"
	"strconv"

	"oip/dpmain/internal/app/domains/apimodel/request"
	"oip/dpmain/internal/app/domains/apimodel/response"
	"oip/dpmain/internal/app/domains/entity/etorder"
	"oip/dpmain/internal/app/pkg/ginx"

	"github.com/gin-gonic/gin"
)

// Create 创建订单接口
// POST /api/v1/orders?wait=10
func (h *OrderHandler) Create(c *gin.Context) {
	waitSeconds := 0
	if waitStr := c.Query("wait"); waitStr != "" {
		if w, err := strconv.Atoi(waitStr); err == nil && w > 0 {
			waitSeconds = w
		}
	}

	var req request.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ginx.BadRequest(c, err.Error())
		return
	}

	shipment := req.ToShipmentEntity()
	order, err := h.orderService.CreateOrder(c.Request.Context(), req.AccountID, req.MerchantOrderNo, shipment, waitSeconds)
	if err != nil {
		log.Printf("[ERROR] create order failed: %v", err)
		ginx.InternalError(c, err.Error())
		return
	}

	if order.Status == etorder.OrderStatusDiagnosed {
		ginx.Success(c, response.FromOrderEntity(order))
	} else if order.Status == etorder.OrderStatusDiagnosing {
		pollURL := fmt.Sprintf("/api/v1/orders/%s", order.ID)
		ginx.Processing(c, order.ID, pollURL)
	} else {
		ginx.Success(c, response.FromOrderEntity(order))
	}
}
