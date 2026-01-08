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

// Create godoc
// @Summary      创建订单
// @Description  创建订单并触发智能诊断（物流费率计算 + 异常检测）
// @Description
// @Description  Smart Wait 机制说明：
// @Description  - 接口会 Hold 10s 等待诊断结果
// @Description  - 10s 内完成诊断：返回 200 OK，包含完整诊断结果
// @Description  - 10s 超时：返回 200 OK，code=3001（Processing），需要通过 poll_url 轮询结果
// @Description
// @Description  订单状态说明：
// @Description  - PENDING: 订单已创建，等待诊断
// @Description  - DIAGNOSING: 诊断进行中
// @Description  - COMPLETED: 诊断完成（成功或失败）
// @Description  - FAILED: 订单处理失败
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        request body request.CreateOrderRequest true "创建订单请求"
// @Success      200 {object} ginx.Response{data=response.OrderResponse} "创建成功"
// @Failure      400 {object} ginx.Response "参数错误"
// @Failure      500 {object} ginx.Response "服务器错误"
// @Security     ApiKeyAuth
// @Router       /orders [post]
func (h *OrderHandler) Create(c *gin.Context) {
	waitSeconds := 0
	if waitStr := c.Query("wait"); waitStr != "" {
		if w, err := strconv.Atoi(waitStr); err == nil && w > 0 {
			waitSeconds = w
		}
	}

	var req request.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ginx.BadRequestWithValidation(c, err)
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
