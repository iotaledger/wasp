package services

import (
	"errors"

	"golang.org/x/exp/maps"

	"github.com/iotaledger/wasp/packages/webapi/v2/models"

	"github.com/iotaledger/wasp/packages/users"

	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/webapi/v2/interfaces"
)

type UserService struct {
	logger *logger.Logger

	userManager *users.UserManager
}

func NewUserService(log *logger.Logger, userManager *users.UserManager) interfaces.UserService {
	return &UserService{
		logger: log,

		userManager: userManager,
	}
}

func (u *UserService) AddUser(username string, password string, permissions []string) {

}

func (u *UserService) UpdateUser(username string, User)

func (u *UserService) DeleteUser(username string) error {
	return u.userManager.RemoveUser(username)
}

func (u *UserService) GetUsers() *[]models.User {
	users := u.userManager.Users()
	userModels := make([]models.User, len(users))

	for i, user := range maps.Values(users) {
		userModels[i] = models.User{
			Username:    user.Name,
			Permissions: user.Permissions,
		}
	}

	return &userModels
}

func (u *UserService) GetUser(username string) (*models.User, error) {
	user, err := u.userManager.User(username)

	if err != nil {
		return nil, err
	}

	return &models.User{
		Username:    user.Name,
		Permissions: user.Permissions,
	}, nil
}

func (v *UserService) getReceipt(chainID *isc.ChainID, requestID isc.RequestID) (*isc.Receipt, *isc.VMError, error) {
	chain := v.chainsProvider().Get(chainID)
	if chain == nil {
		return nil, nil, errors.New("chain does not exist")
	}

	receipt, err := chain.GetRequestReceipt(requestID)
	if err != nil {
		return nil, nil, err
	}

	resolvedError, err := chain.ResolveError(receipt.Error)
	if err != nil {
		return nil, nil, err
	}

	receiptData := receipt.ToISCReceipt(resolvedError)

	return receiptData, resolvedError, nil
}
