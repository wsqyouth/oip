package domains

import (
	"context"

	"oip/dpsync/internal/business/order/diagnose"
	"oip/dpsync/internal/framework"
	"oip/dpsync/pkg/lmstfy"
)

// HandlerFactory Handler 构造函数类型
type HandlerFactory func(
	ctx context.Context,
	baseHandler *framework.BaseHandler,
	lmstfyClient *lmstfy.Client,
	callbackQueue string,
) (framework.BusinessHandler, error)

// HandlerMap 路由表（ActionType → Handler 映射）
var HandlerMap = map[string]HandlerFactory{
	"order_diagnose": diagnose.NewDiagnoseHandler,

	// 未来扩展示例：
	// "order_risk_check": order_risk.NewRiskCheckHandler,
}
