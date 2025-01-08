//go:generate mockgen -package=mocks -destination ../mocks/user_service.go liquiswiss/internal/service/user_service IUserService
package user_service

import (
	"liquiswiss/internal/service/db_service"
	"liquiswiss/pkg/models"
)

type IUserService interface {
	GetCurrentOrganisation(userID int64) (*models.Organisation, error)
}

type UserServiceService struct {
	dbService db_service.IDatabaseService
}

func NewUserServiceService(s *db_service.IDatabaseService) IUserService {
	return &UserServiceService{
		dbService: *s,
	}
}

func (u *UserServiceService) GetCurrentOrganisation(userID int64) (*models.Organisation, error) {
	profile, err := u.dbService.GetProfile(userID)
	if err != nil {
		return nil, err
	}

	organisation, err := u.dbService.GetOrganisation(userID, profile.CurrentOrganisationID)
	if err != nil {
		return nil, err
	}

	return organisation, nil
}
