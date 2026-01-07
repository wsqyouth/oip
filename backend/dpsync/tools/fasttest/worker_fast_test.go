package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"oip/dpsync/internal/business"
	"oip/dpsync/pkg/config"
	"oip/dpsync/pkg/infra/mysql"
	"oip/dpsync/pkg/infra/redis"
)

var (
	configPath   = flag.String("config", "./config/worker.yaml", "配置文件路径")
	testcasePath = flag.String("testcase", "./internal/domains/handlers/order/diagnose/testcase/diagnose.json", "测试用例路径")
	skipDB       = flag.Bool("skip-db", false, "跳过数据库操作（仅测试业务逻辑）")
)

// TestCase 测试用例结构
type TestCase struct {
	OrderID   string `json:"order_id"`
	AccountID int64  `json:"account_id"`
}

func main() {
	flag.Parse()

	fmt.Println("========================================")
	fmt.Println("  FastTest - DPSYNC Worker 快速测试工具")
	fmt.Println("========================================")

	// 1. 加载配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Printf("❌ Failed to load config: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Config loaded: %s\n", cfg.App.Name)

	// 2. 加载测试用例
	testCases, err := loadTestCases(*testcasePath)
	if err != nil {
		fmt.Printf("❌ Failed to load test cases: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Loaded %d test cases from %s\n", len(testCases), *testcasePath)

	// 3. 初始化依赖（根据 skip-db 参数决定）
	var diagnosisService *business.DiagnosisService
	if *skipDB {
		fmt.Println("⚠️  Skip-DB mode: Database and Redis operations disabled")
		// 只测试业务逻辑，不连接数据库和 Redis
		diagnosisService = nil
	} else {
		// 完整模式：初始化数据库和 Redis
		orderDAO, err := mysql.NewOrderDAO(cfg.MySQL.DSN)
		if err != nil {
			fmt.Printf("❌ Failed to create OrderDAO: %v\n", err)
			os.Exit(1)
		}
		defer orderDAO.Close()

		redisPubSub, err := redis.NewPubSub(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
		if err != nil {
			fmt.Printf("❌ Failed to create Redis PubSub: %v\n", err)
			os.Exit(1)
		}
		defer redisPubSub.Close()

		diagnosisService = business.NewDiagnosisService(
			orderDAO,
			redisPubSub,
			"order_diagnosis_complete",
		)
		fmt.Println("✅ Database and Redis initialized")
	}

	// 4. 执行测试用例
	fmt.Println("\n========================================")
	fmt.Println("  Running Test Cases")
	fmt.Println("========================================")

	successCount := 0
	failureCount := 0

	for i, tc := range testCases {
		fmt.Printf("\n[Test %d/%d] OrderID=%s, AccountID=%d\n", i+1, len(testCases), tc.OrderID, tc.AccountID)
		fmt.Println("----------------------------------------")

		startTime := time.Now()

		if *skipDB {
			// Skip-DB 模式：只测试 CompositeHandler
			err = runTestCaseSkipDB(tc)
		} else {
			// 完整模式：测试完整诊断流程
			err = runTestCaseFull(diagnosisService, tc)
		}

		duration := time.Since(startTime)

		if err != nil {
			fmt.Printf("❌ FAILED: %v\n", err)
			fmt.Printf("⏱️  Duration: %v\n", duration)
			failureCount++
		} else {
			fmt.Printf("✅ PASSED\n")
			fmt.Printf("⏱️  Duration: %v\n", duration)
			successCount++
		}
	}

	// 5. 输出测试汇总
	fmt.Println("\n========================================")
	fmt.Println("  Test Summary")
	fmt.Println("========================================")
	fmt.Printf("Total: %d\n", len(testCases))
	fmt.Printf("Passed: %d ✅\n", successCount)
	fmt.Printf("Failed: %d ❌\n", failureCount)

	if failureCount > 0 {
		os.Exit(1)
	}
}

// loadTestCases 从 JSON 文件加载测试用例
func loadTestCases(path string) ([]TestCase, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read testcase file: %w", err)
	}

	var testCases []TestCase
	if err := json.Unmarshal(data, &testCases); err != nil {
		return nil, fmt.Errorf("failed to unmarshal testcase: %w", err)
	}

	return testCases, nil
}

// runTestCaseSkipDB 运行测试用例（跳过数据库，仅测试业务逻辑）
func runTestCaseSkipDB(tc TestCase) error {
	ctx := context.Background()

	// 创建 CompositeHandler
	compositeHandler := business.NewCompositeHandler()

	// 执行诊断
	input := &business.DiagnoseInput{
		OrderID:   tc.OrderID,
		AccountID: tc.AccountID,
	}

	result, err := compositeHandler.Diagnose(ctx, input)
	if err != nil {
		return fmt.Errorf("diagnosis failed: %w", err)
	}

	// 打印诊断结果
	fmt.Printf("  Diagnosis Items: %d\n", len(result.Items))
	for _, item := range result.Items {
		fmt.Printf("    - Type=%s, Status=%s\n", item.Type, item.Status)
		if item.Error != "" {
			fmt.Printf("      Error: %s\n", item.Error)
		}
	}

	return nil
}

// runTestCaseFull 运行测试用例（完整模式：诊断 + 数据库 + Redis）
func runTestCaseFull(diagnosisService *business.DiagnosisService, tc TestCase) error {
	ctx := context.Background()

	// 执行完整诊断流程
	result := diagnosisService.ExecuteDiagnosis(ctx, tc.OrderID, tc.AccountID)
	if !result.Success {
		return fmt.Errorf("diagnosis failed: %w", result.Error)
	}

	// 打印诊断结果
	fmt.Printf("  Diagnosis Items: %d\n", len(result.Data.Items))
	for _, item := range result.Data.Items {
		fmt.Printf("    - Type=%s, Status=%s\n", item.Type, item.Status)
		if item.Error != "" {
			fmt.Printf("      Error: %s\n", item.Error)
		}
	}
	fmt.Println("  ✓ Database updated")
	fmt.Println("  ✓ Redis notification sent")

	return nil
}
