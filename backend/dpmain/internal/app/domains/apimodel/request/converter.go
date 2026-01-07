package request

import "oip/dpmain/internal/app/domains/entity/etorder"

// ToShipmentEntity 将 Request DTO 转换为领域对象
func (r *CreateOrderRequest) ToShipmentEntity() *etorder.Shipment {
	return &etorder.Shipment{
		ShipFrom: toAddressEntity(r.Shipment.ShipFrom),
		ShipTo:   toAddressEntity(r.Shipment.ShipTo),
		Parcels:  toParcelsEntity(r.Shipment.Parcels),
	}
}

func toAddressEntity(dto *Address) *etorder.Address {
	if dto == nil {
		return nil
	}
	return &etorder.Address{
		ContactName: dto.ContactName,
		CompanyName: dto.CompanyName,
		Street1:     dto.Street1,
		Street2:     dto.Street2,
		City:        dto.City,
		State:       dto.State,
		PostalCode:  dto.PostalCode,
		Country:     dto.Country,
		Phone:       dto.Phone,
		Email:       dto.Email,
	}
}

func toParcelsEntity(dtos []*Parcel) []*etorder.Parcel {
	parcels := make([]*etorder.Parcel, 0, len(dtos))
	for _, dto := range dtos {
		parcels = append(parcels, &etorder.Parcel{
			Weight:    toWeightEntity(dto.Weight),
			Dimension: toDimensionEntity(dto.Dimension),
			Items:     toItemsEntity(dto.Items),
		})
	}
	return parcels
}

func toWeightEntity(dto *Weight) *etorder.Weight {
	if dto == nil {
		return nil
	}
	return &etorder.Weight{
		Value: dto.Value,
		Unit:  dto.Unit,
	}
}

func toDimensionEntity(dto *Dimension) *etorder.Dimension {
	if dto == nil {
		return nil
	}
	return &etorder.Dimension{
		Width:  dto.Width,
		Height: dto.Height,
		Depth:  dto.Depth,
		Unit:   dto.Unit,
	}
}

func toItemsEntity(dtos []*Item) []*etorder.Item {
	items := make([]*etorder.Item, 0, len(dtos))
	for _, dto := range dtos {
		items = append(items, &etorder.Item{
			Description: dto.Description,
			Quantity:    dto.Quantity,
			Price:       toMoneyEntity(dto.Price),
			SKU:         dto.SKU,
			Weight:      toWeightEntity(dto.Weight),
		})
	}
	return items
}

func toMoneyEntity(dto *Money) *etorder.Money {
	if dto == nil {
		return nil
	}
	return &etorder.Money{
		Amount:   dto.Amount,
		Currency: dto.Currency,
	}
}
