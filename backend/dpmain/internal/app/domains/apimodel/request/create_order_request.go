package request

// CreateOrderRequest 创建订单请求（DTO）
type CreateOrderRequest struct {
	AccountID       int64     `json:"account_id" binding:"required"`
	MerchantOrderNo string    `json:"merchant_order_no" binding:"required"`
	Shipment        *Shipment `json:"shipment" binding:"required"`
}

// Shipment 货件信息（DTO）
type Shipment struct {
	ShipFrom *Address  `json:"ship_from" binding:"required"`
	ShipTo   *Address  `json:"ship_to" binding:"required"`
	Parcels  []*Parcel `json:"parcels" binding:"required"`
}

// Address 地址（DTO）
type Address struct {
	ContactName string `json:"contact_name" binding:"required"`
	CompanyName string `json:"company_name"`
	Street1     string `json:"street1" binding:"required"`
	Street2     string `json:"street2"`
	City        string `json:"city" binding:"required"`
	State       string `json:"state"`
	PostalCode  string `json:"postal_code" binding:"required"`
	Country     string `json:"country" binding:"required"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
}

// Parcel 包裹（DTO）
type Parcel struct {
	Weight    *Weight    `json:"weight" binding:"required"`
	Dimension *Dimension `json:"dimension"`
	Items     []*Item    `json:"items" binding:"required"`
}

// Weight 重量（DTO）
type Weight struct {
	Value float64 `json:"value" binding:"required"`
	Unit  string  `json:"unit" binding:"required"`
}

// Dimension 尺寸（DTO）
type Dimension struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
	Depth  float64 `json:"depth"`
	Unit   string  `json:"unit"`
}

// Item 商品（DTO）
type Item struct {
	Description string  `json:"description" binding:"required"`
	Quantity    int     `json:"quantity" binding:"required"`
	Price       *Money  `json:"price" binding:"required"`
	SKU         string  `json:"sku"`
	Weight      *Weight `json:"weight"`
}

// Money 金额（DTO）
type Money struct {
	Amount   float64 `json:"amount" binding:"required"`
	Currency string  `json:"currency" binding:"required"`
}
