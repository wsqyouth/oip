package etorder

import (
	"errors"
	"time"
)

// 错误定义
var (
	ErrInvalidOrderID         = errors.New("order ID cannot be empty")
	ErrInvalidAccountID       = errors.New("invalid account ID")
	ErrInvalidMerchantOrderNo = errors.New("merchant order number cannot be empty")
	ErrInvalidShipment        = errors.New("invalid shipment data")
	ErrNilDiagnoseResult      = errors.New("diagnose result cannot be nil")
)

// Order 订单聚合根（领域对象）
type Order struct {
	ID              string          // 订单ID (UUID)
	AccountID       int64           // 账户ID
	MerchantOrderNo string          // 商户订单号
	Shipment        *Shipment       // 货件信息
	Status          OrderStatus     // 订单状态
	DiagnoseResult  *DiagnoseResult // 诊断结果
	CreatedAt       time.Time       // 创建时间
	UpdatedAt       time.Time       // 更新时间
}

// OrderStatus 订单状态
type OrderStatus string

const (
	OrderStatusDiagnosing OrderStatus = "DIAGNOSING"
	OrderStatusDiagnosed  OrderStatus = "DIAGNOSED"
	OrderStatusFailed     OrderStatus = "FAILED"
)

// Shipment 货件信息（值对象）
type Shipment struct {
	ShipFrom *Address
	ShipTo   *Address
	Parcels  []*Parcel
}

// Address 地址（值对象）
type Address struct {
	ContactName string
	CompanyName string
	Street1     string
	Street2     string
	City        string
	State       string
	PostalCode  string
	Country     string
	Phone       string
	Email       string
}

// Parcel 包裹（值对象）
type Parcel struct {
	Weight    *Weight
	Dimension *Dimension
	Items     []*Item
}

// Weight 重量（值对象）
type Weight struct {
	Value float64
	Unit  string
}

// Dimension 尺寸（值对象）
type Dimension struct {
	Width  float64
	Height float64
	Depth  float64
	Unit   string
}

// Item 商品（值对象）
type Item struct {
	Description string
	Quantity    int
	Price       *Money
	SKU         string
	Weight      *Weight
}

// Money 金额（值对象）
type Money struct {
	Amount   float64
	Currency string
}

// DiagnoseResult 诊断结果（值对象）
type DiagnoseResult struct {
	Items []*DiagnoseItem
}

// DiagnoseItem 单个诊断项
type DiagnoseItem struct {
	Type     string
	Status   string
	DataJSON interface{}
	Error    string
}

// NewOrder 创建订单（工厂方法）
func NewOrder(id string, accountID int64, merchantOrderNo string, shipment *Shipment) (*Order, error) {
	// 业务规则校验
	if id == "" {
		return nil, ErrInvalidOrderID
	}
	if accountID <= 0 {
		return nil, ErrInvalidAccountID
	}
	if merchantOrderNo == "" {
		return nil, ErrInvalidMerchantOrderNo
	}
	if shipment == nil {
		return nil, ErrInvalidShipment
	}

	return &Order{
		ID:              id,
		AccountID:       accountID,
		MerchantOrderNo: merchantOrderNo,
		Shipment:        shipment,
		Status:          OrderStatusDiagnosing,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}, nil
}

// UpdateDiagnoseResult 更新诊断结果（领域行为）
func (o *Order) UpdateDiagnoseResult(result *DiagnoseResult) error {
	if result == nil {
		return ErrNilDiagnoseResult
	}
	o.DiagnoseResult = result
	o.Status = OrderStatusDiagnosed
	o.UpdatedAt = time.Now()
	return nil
}

// MarkAsFailed 标记为失败（领域行为）
func (o *Order) MarkAsFailed() {
	o.Status = OrderStatusFailed
	o.UpdatedAt = time.Now()
}
