package domains

import (
	"oip/dpsync/internal/domains/common"
	"oip/dpsync/internal/domains/handlers/order/diagnose"
)

// HandlerMap 路由表（ActionType → Handler 映射）
var HandlerMap = map[string]common.HandlerServProc{
	// 订单诊断
	"order_diagnose": diagnose.NewDiagnoseHandler,

	// 未来扩展示例：
	// "order_risk_check": order_risk.NewRiskCheckHandler,
	// "order_compliance": order_compliance.NewComplianceHandler,
}
