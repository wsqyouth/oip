package mddiagnosis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"oip/common/model"
	"oip/dpmain/internal/app/domains/entity/etorder"
	"oip/dpmain/internal/app/infra/mq/lmstfy"
	"oip/dpmain/internal/app/infra/persistence/redis"
)

// DiagnosisModule 诊断模块
// 职责：
// 1. 组装 Lmstfy 和 Redis 客户端
// 2. 包含诊断相关的业务逻辑（消息格式构造、频道命名规则）
type DiagnosisModule struct {
	lmstfyClient *lmstfy.Client
	redisClient  *redis.PubSubClient
	queueName    string
}

// NewDiagnosisModule 创建诊断模块实例
func NewDiagnosisModule(lmstfyClient *lmstfy.Client, redisClient *redis.PubSubClient, queueName string) *DiagnosisModule {
	return &DiagnosisModule{
		lmstfyClient: lmstfyClient,
		redisClient:  redisClient,
		queueName:    queueName,
	}
}

// PublishDiagnoseJob 发布订单诊断任务到队列
// 业务逻辑：
// 1. 构造标准化消息格式（包含 RequestID, ActionType, OrgID 等）
// 2. 将 Shipment 转换为 map 格式（避免 dpsync 查询 DB）
func (m *DiagnosisModule) PublishDiagnoseJob(ctx context.Context, order *etorder.Order) error {
	// 业务逻辑：将 Shipment 结构体转换为 map
	var shipmentMap map[string]interface{}
	shipmentJSON, err := json.Marshal(order.Shipment)
	if err != nil {
		return fmt.Errorf("marshal shipment failed: %w", err)
	}
	if err := json.Unmarshal(shipmentJSON, &shipmentMap); err != nil {
		return fmt.Errorf("unmarshal shipment to map failed: %w", err)
	}

	// 业务逻辑：构造标准化消息格式
	message := model.OrderDiagnoseJob{
		Payload: model.OrderDiagnosePayload{
			Data: model.OrderDiagnoseData{
				RequestID:  uuid.New().String(), // 生成请求 ID 用于全链路追踪
				OrgID:      "0",                 // MVP 固定值
				ActionType: "order_diagnose",
				ID:         order.ID,
				Data: model.OrderDiagnoseBusinessData{
					OrderID:         order.ID,
					AccountID:       order.AccountID,
					MerchantOrderNo: order.MerchantOrderNo,
					Shipment:        shipmentMap, // 传递完整的 shipment 数据（map 格式）
				},
			},
		},
	}

	// 调用基础设施层
	return m.lmstfyClient.Publish(ctx, m.queueName, message)
}

// WaitForDiagnosisResult 等待诊断结果（Smart Wait）
// 业务逻辑：
// 1. 知道订阅哪个频道（业务约定：diagnosis:result:{orderID}）
// 2. 解析诊断结果为领域对象
func (m *DiagnosisModule) WaitForDiagnosisResult(ctx context.Context, orderID string, timeout time.Duration) (*etorder.DiagnoseResult, error) {
	// 业务逻辑：频道命名规则
	channel := fmt.Sprintf("diagnosis:result:%s", orderID)

	// 调用基础设施层
	payload, err := m.redisClient.Subscribe(ctx, channel, timeout)
	if err != nil {
		return nil, err
	}

	// 业务逻辑：反序列化为领域对象
	var result etorder.DiagnoseResult
	if err := json.Unmarshal([]byte(payload), &result); err != nil {
		return nil, err
	}

	return &result, nil
}
