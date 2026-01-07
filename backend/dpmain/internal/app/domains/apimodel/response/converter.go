package response

import (
	"oip/dpmain/internal/app/domains/entity/etaccount"
	"oip/dpmain/internal/app/domains/entity/etorder"
)

// FromOrderEntity 从领域对象转换为响应 DTO
func FromOrderEntity(order *etorder.Order) *OrderResponse {
	resp := &OrderResponse{
		ID:              order.ID,
		AccountID:       order.AccountID,
		MerchantOrderNo: order.MerchantOrderNo,
		Status:          string(order.Status),
		CreatedAt:       order.CreatedAt,
		UpdatedAt:       order.UpdatedAt,
	}

	if order.DiagnoseResult != nil {
		resp.Diagnosis = fromDiagnosisEntity(order.DiagnoseResult)
	}

	return resp
}

func fromDiagnosisEntity(entity *etorder.DiagnoseResult) *DiagnosisResult {
	if entity == nil {
		return nil
	}

	items := make([]*DiagnosisItem, 0, len(entity.Items))
	for _, item := range entity.Items {
		items = append(items, &DiagnosisItem{
			Type:     item.Type,
			Status:   item.Status,
			DataJSON: item.DataJSON,
			Error:    item.Error,
		})
	}

	return &DiagnosisResult{Items: items}
}

// FromAccountEntity 从领域对象转换为响应 DTO
func FromAccountEntity(account *etaccount.Account) *AccountResponse {
	return &AccountResponse{
		ID:        account.ID,
		Name:      account.Name,
		Email:     account.Email,
		CreatedAt: account.CreatedAt,
	}
}
