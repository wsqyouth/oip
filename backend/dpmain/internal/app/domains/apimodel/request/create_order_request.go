package request

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	AccountID       int64     `json:"account_id" binding:"required" example:"1"`
	MerchantOrderNo string    `json:"merchant_order_no" binding:"required" example:"ORD-20240101-001"`
	Shipment        *Shipment `json:"shipment" binding:"required"`
}

// Shipment 货件信息
type Shipment struct {
	ShipFrom *Address  `json:"ship_from" binding:"required"`
	ShipTo   *Address  `json:"ship_to" binding:"required"`
	Parcels  []*Parcel `json:"parcels" binding:"required"`
}

// Address 地址信息
type Address struct {
	ContactName string `json:"contact_name" binding:"required" example:"John Doe"`
	CompanyName string `json:"company_name" example:"ACME Corp"`
	Street1     string `json:"street1" binding:"required" example:"123 Main St"`
	Street2     string `json:"street2" example:"Suite 100"`
	City        string `json:"city" binding:"required" example:"San Francisco"`
	State       string `json:"state" example:"CA"`
	PostalCode  string `json:"postal_code" binding:"required" example:"94102"`
	Country     string `json:"country" binding:"required" example:"USA"`
	Phone       string `json:"phone" example:"+1-415-555-0100"`
	Email       string `json:"email" example:"john@example.com"`
}

// Parcel 包裹信息
type Parcel struct {
	Weight    *Weight    `json:"weight" binding:"required"`
	Dimension *Dimension `json:"dimension"`
	Items     []*Item    `json:"items" binding:"required"`
}

// Weight 重量信息
type Weight struct {
	Value float64 `json:"value" binding:"required" example:"1.5"`
	Unit  string  `json:"unit" binding:"required" example:"kg"`
}

// Dimension 尺寸信息
type Dimension struct {
	Width  float64 `json:"width" example:"10.0"`
	Height float64 `json:"height" example:"20.0"`
	Depth  float64 `json:"depth" example:"15.0"`
	Unit   string  `json:"unit" example:"cm"`
}

// Item 商品信息
type Item struct {
	Description string  `json:"description" binding:"required" example:"T-Shirt"`
	Quantity    int     `json:"quantity" binding:"required" example:"2"`
	Price       *Money  `json:"price" binding:"required"`
	SKU         string  `json:"sku" example:"TSH-001"`
	Weight      *Weight `json:"weight"`
}

// Money 金额信息
type Money struct {
	Amount   float64 `json:"amount" binding:"required" example:"19.99"`
	Currency string  `json:"currency" binding:"required" example:"USD"`
}
