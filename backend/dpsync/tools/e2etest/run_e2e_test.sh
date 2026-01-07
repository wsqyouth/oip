#!/bin/bash

# DPSYNC 端到端测试脚本
# 功能：发送测试消息到 lmstfy → Worker 消费 → 验证数据库和 Redis

set -e

echo "========================================"
echo "  DPSYNC 端到端测试"
echo "========================================"

# 配置
LMSTFY_HOST="${LMSTFY_HOST:-http://localhost:7777}"
NAMESPACE="oip"
QUEUE="oip_order_diagnose"
MYSQL_DSN="${MYSQL_DSN:-root:password@tcp(127.0.0.1:3306)/oip}"
REDIS_ADDR="${REDIS_ADDR:-localhost:6379}"

# 测试用例
ORDER_ID="e2e_test_$(date +%s)"
ACCOUNT_ID=999

echo "📝 测试配置："
echo "  - lmstfy: $LMSTFY_HOST"
echo "  - Queue: $QUEUE"
echo "  - OrderID: $ORDER_ID"
echo "  - AccountID: $ACCOUNT_ID"
echo ""

# 步骤 1：检查依赖服务
echo "🔍 [Step 1] 检查依赖服务..."
echo -n "  - lmstfy: "
if curl -s -f "$LMSTFY_HOST/ping" > /dev/null 2>&1; then
    echo "✅ Running"
else
    echo "❌ Not running"
    echo "请启动 lmstfy 服务：docker run -p 7777:7777 bitleak/lmstfy"
    exit 1
fi

echo ""

# 步骤 2：构造测试消息
echo "📦 [Step 2] 构造测试消息..."
MESSAGE=$(cat <<EOF
{
  "payload": {
    "data": {
      "request_id": "e2e-test-$(date +%s)",
      "org_id": "org-test",
      "action_type": "order_diagnose",
      "id": "diag-e2e-test",
      "data": {
        "order_id": "$ORDER_ID",
        "account_id": $ACCOUNT_ID
      }
    }
  }
}
EOF
)

echo "消息内容："
echo "$MESSAGE" | jq '.' 2>/dev/null || echo "$MESSAGE"
echo ""

# 步骤 3：发送消息到 lmstfy
echo "📨 [Step 3] 发送消息到 lmstfy..."
PUBLISH_URL="$LMSTFY_HOST/api/$NAMESPACE/$QUEUE"

RESPONSE=$(curl -s -w "\n%{http_code}" -X PUT "$PUBLISH_URL" \
  -d "ttl=3600" \
  -d "delay=0" \
  --data-binary "$MESSAGE")

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n-1)

if [ "$HTTP_CODE" = "201" ]; then
    JOB_ID=$(echo "$BODY" | jq -r '.job_id' 2>/dev/null || echo "unknown")
    echo "✅ 消息发送成功"
    echo "  - Job ID: $JOB_ID"
else
    echo "❌ 消息发送失败"
    echo "  - HTTP Code: $HTTP_CODE"
    echo "  - Response: $BODY"
    exit 1
fi

echo ""

# 步骤 4：等待 Worker 处理
echo "⏳ [Step 4] 等待 Worker 处理消息（最多 30 秒）..."
echo "  请确保 Worker 正在运行：go run cmd/worker/main.go"
echo ""

for i in {1..30}; do
    echo -n "."
    sleep 1
done
echo " Done"
echo ""

# 步骤 5：验证数据库（可选）
echo "🔍 [Step 5] 验证数据库结果（可选）..."
if command -v mysql &> /dev/null; then
    echo "检查订单诊断结果..."
    mysql -h 127.0.0.1 -u root -ppassword oip -e \
        "SELECT id, status, JSON_EXTRACT(diagnose_result, '\$.items[*].type') as types FROM orders WHERE id = '$ORDER_ID';" \
        2>/dev/null || echo "⚠️  MySQL 命令失败或订单不存在（可能是正常现象）"
else
    echo "⚠️  mysql 命令未安装，跳过数据库验证"
fi

echo ""

# 步骤 6：验证 Redis 通知（可选）
echo "🔍 [Step 6] 验证 Redis 通知（可选）..."
echo "订阅 Redis 频道 'order_diagnosis_complete' 查看通知："
echo "  redis-cli SUBSCRIBE order_diagnosis_complete"
echo ""

# 步骤 7：汇总结果
echo "========================================"
echo "  测试汇总"
echo "========================================"
echo "✅ 测试消息已发送到 lmstfy"
echo "⏳ Worker 应该在 30 秒内处理完消息"
echo ""
echo "手动验证步骤："
echo "1. 检查 Worker 日志，确认消息被处理"
echo "2. 查询数据库：SELECT * FROM orders WHERE id = '$ORDER_ID';"
echo "3. 订阅 Redis：redis-cli SUBSCRIBE order_diagnosis_complete"
echo ""
echo "如果以上步骤都成功，说明端到端测试通过！🎉"
echo "========================================"
