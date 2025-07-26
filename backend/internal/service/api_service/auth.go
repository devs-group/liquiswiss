package api_service

import (
	"golang.org/x/crypto/bcrypt"
	"liquiswiss/pkg/auth"
	"liquiswiss/pkg/logger"
	"liquiswiss/pkg/models"
	"liquiswiss/pkg/utils"
	"time"
)

func (a *APIService) Login(payload models.Login, deviceName string, existingRefreshToken string) (*models.User, *string, *time.Time, *string, *time.Time, error) {
	loginUser, err := a.dbService.GetUserPasswordByEMail(payload.Email)
	if err != nil {
		logger.Logger.Error(err)
		return nil, nil, nil, nil, nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(loginUser.Password), []byte(payload.Password))
	if err != nil {
		logger.Logger.Error(err)
		return nil, nil, nil, nil, nil, err
	}

	user, err := a.dbService.GetProfile(loginUser.ID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, nil, nil, nil, nil, err
	}

	accessToken, accessExpirationTime, _, err := auth.GenerateAccessToken(*user)
	if err != nil {
		logger.Logger.Error(err)
		return nil, nil, nil, nil, nil, err
	}

	refreshToken, tokenId, refreshExpirationTime, err := auth.GenerateRefreshToken(*user)
	if err != nil {
		logger.Logger.Error(err)
		return nil, nil, nil, nil, nil, err
	}

	// Clear the pre-existing refreshToken first
	a.clearRefreshTokenFromDatabase(existingRefreshToken)

	// Store the refresh token in the database
	err = a.dbService.StoreRefreshTokenID(loginUser.ID, tokenId, refreshExpirationTime, deviceName)
	if err != nil {
		logger.Logger.Error(err)
		return nil, nil, nil, nil, nil, err
	}

	return user, &accessToken, &accessExpirationTime, &refreshToken, &refreshExpirationTime, nil
}

func (a *APIService) Logout(existingRefreshToken string) {
	a.clearRefreshTokenFromDatabase(existingRefreshToken)
}

func (a *APIService) ForgotPassword(payload models.ForgotPassword, code string) error {
	hasCreated, err := a.dbService.CreateResetPassword(payload.Email, code, utils.ResetPasswordDelay)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}
	if hasCreated {
		err := a.sendgridAdapter.SendPasswordResetMail(payload.Email, code)
		if err != nil {
			logger.Logger.Error(err)

			err := a.dbService.DeleteResetPassword(payload.Email)
			if err != nil {
				logger.Logger.Error(err)
			}
			return err
		}
	} else {
		logger.Logger.Infof("Skipped creating password reset")
	}
	return nil
}

func (a *APIService) ResetPassword(payload models.ResetPassword) error {
	_, err := a.dbService.ValidateResetPassword(payload.Email, payload.Code, utils.ResetPasswordDelay)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}

	encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), 12)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}

	err = a.dbService.ResetPassword(string(encryptedPassword), payload.Email)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}

	err = a.dbService.DeleteResetPassword(payload.Email)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}

	return nil
}

func (a *APIService) CheckResetPasswordCode(payload models.CheckResetPasswordCode) error {
	_, err := a.dbService.ValidateResetPassword(payload.Email, payload.Code, utils.ResetPasswordDelay)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}
	return nil
}

func (a *APIService) CreateRegistration(payload models.CreateRegistration, code string) (int64, error) {
	registrationID, err := a.dbService.CreateRegistration(payload.Email, code)
	if err != nil {
		logger.Logger.Error(err)
		return 0, err
	}

	// Make sure a registration can't exist if the email already exists for a user
	if registrationID == 0 {
		return 0, nil
	}

	err = a.sendgridAdapter.SendRegistrationMail(payload.Email, code)
	if err != nil {
		logger.Logger.Error(err)

		err2 := a.DeleteRegistration(registrationID, code)
		if err2 != nil {
			logger.Logger.Error(err2)
		}
		return 0, err
	}

	return registrationID, nil
}

func (a *APIService) CheckRegistrationCode(payload models.CheckRegistrationCode, validity time.Duration) (int64, error) {
	registrationID, err := a.dbService.ValidateRegistration(payload.Email, payload.Code, validity)
	if err != nil {
		logger.Logger.Error(err)
		return 0, err
	}
	return registrationID, nil
}

func (a *APIService) FinishRegistration(payload models.FinishRegistration, deviceName string, validity time.Duration) (*models.User, *string, *time.Time, *string, *time.Time, error) {
	registrationId, err := a.CheckRegistrationCode(models.CheckRegistrationCode{
		Email: payload.Email,
		Code:  payload.Code,
	}, validity)
	if err != nil {
		logger.Logger.Error(err)
		return nil, nil, nil, nil, nil, err
	}

	encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), 12)
	if err != nil {
		logger.Logger.Error(err)
		return nil, nil, nil, nil, nil, err
	}

	userId, err := a.dbService.CreateUser(payload.Email, string(encryptedPassword))
	if err != nil {
		logger.Logger.Error(err)
		return nil, nil, nil, nil, nil, err
	}

	// Every new user gets an organisation assigned automatically
	organisationID, err := a.dbService.CreateOrganisation("Meine Organisation")
	if err != nil {
		logger.Logger.Error(err)
		return nil, nil, nil, nil, nil, err
	}

	// We explicity set it as the default organisation which can only be deleted along with the users account
	err = a.dbService.AssignUserToOrganisation(userId, organisationID, "owner", true)
	if err != nil {
		logger.Logger.Error(err)
		return nil, nil, nil, nil, nil, err
	}

	// Set the new default organisation as the current one
	err = a.dbService.SetUserCurrentOrganisation(userId, organisationID)
	if err != nil {
		logger.Logger.Error(err)
		return nil, nil, nil, nil, nil, err
	}

	// Delete registration since we have the user now
	err = a.dbService.DeleteRegistration(registrationId, payload.Email)
	if err != nil {
		logger.Logger.Error(err)
		return nil, nil, nil, nil, nil, err
	}

	user, err := a.dbService.GetProfile(userId)
	if err != nil {
		logger.Logger.Error(err)
		return nil, nil, nil, nil, nil, err
	}

	accessToken, accessExpirationTime, _, err := auth.GenerateAccessToken(*user)
	if err != nil {
		logger.Logger.Error(err)
		return nil, nil, nil, nil, nil, err
	}

	refreshToken, tokenId, refreshExpirationTime, err := auth.GenerateRefreshToken(*user)
	if err != nil {
		logger.Logger.Error(err)
		return nil, nil, nil, nil, nil, err
	}

	// Store the refresh token in the database
	err = a.dbService.StoreRefreshTokenID(user.ID, tokenId, refreshExpirationTime, deviceName)
	if err != nil {
		logger.Logger.Error(err)
		return nil, nil, nil, nil, nil, err
	}

	return user, &accessToken, &accessExpirationTime, &refreshToken, &refreshExpirationTime, nil
}

func (a *APIService) DeleteRegistration(registrationID int64, email string) error {
	err := a.dbService.DeleteRegistration(registrationID, email)
	if err != nil {
		logger.Logger.Error(err)
		return err
	}
	return nil
}

func (a *APIService) clearRefreshTokenFromDatabase(refreshToken string) {
	if refreshToken != "" {
		// Verify the refresh token to get its claims (e.g., the tokenID and userID)
		refreshClaims, err := auth.VerifyToken(refreshToken)
		if err == nil {
			// Delete the refresh token from the database
			err = a.dbService.DeleteRefreshToken(refreshClaims.ID, refreshClaims.UserID)
			if err != nil {
				logger.Logger.Error(err)
			}
		}
	}
}
