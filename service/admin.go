package service

import (
	"net/http"
	"urllite/store"
	"urllite/types"
	"urllite/types/dtos"
)

type AdminService interface {
	Dashboard() (*dtos.AdminDashboardDTO, *types.ApplicationError)
}

type adminService struct {
	store store.Store
}

func NewAdminService() AdminService {
	store := store.NewStore()
	return adminService{store: store}
}

func (s adminService) Dashboard() (*dtos.AdminDashboardDTO, *types.ApplicationError) {
	dashboard, err := s.store.AdminDashboard()
	if err != nil {
		return nil, &types.ApplicationError{
			Message:        "Unable to get dashboard data",
			HttpStatusCode: http.StatusInternalServerError,
			Err:            err,
		}
	}
	return dashboard, nil
}
