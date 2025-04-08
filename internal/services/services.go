package services

import "kadane.xyz/go-backend/v2/internal/database/dbaccessors"

type Service interface {
	AccountService
}

type ServiceAccessor struct {
	AccountService
}

func NewServiceAccessor(dbaccessor dbaccessors.AccountAccessor) ServiceAccessor {
	return ServiceAccessor{
		AccountService: NewAccountService(dbaccessor),
	}
}
