package services

import (
	"context"
	"fmt"
	"log/slog"

	openfgaclient "github.com/openfga/go-sdk/client"
	"github.com/pkg/errors"
)

type AuthorizationService interface {
	CreateUserAccess(ctx context.Context, userID string, entityID string, action string) error
	ReadUserAccess(ctx context.Context, userID string, entityID string, action string) (bool, error)
	ListObjectsUserCanAccess(ctx context.Context, userID string, entityType string, action string) ([]string, error)
	DeleteUserAccess(ctx context.Context, userID string, entityID string, action string) error
}

func NewOpenFGAService(logger *slog.Logger, fgaApiUrl string, fgaStoreID string, authorizationModleID string) (*OpenFGAService, error) {
	fgaClient, err := openfgaclient.NewSdkClient(&openfgaclient.ClientConfiguration{
		ApiUrl:               fgaApiUrl,
		StoreId:              fgaStoreID,
		AuthorizationModelId: authorizationModleID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create FGA client")
	}
	return &OpenFGAService{
		openfgaClient: fgaClient,
		logger:        logger,
	}, nil
}

var _ AuthorizationService = &OpenFGAService{}

type OpenFGAService struct {
	openfgaClient *openfgaclient.OpenFgaClient
	logger        *slog.Logger
}

func (o OpenFGAService) CreateUserAccess(ctx context.Context, userID string, entityID string, action string) error {
	body := openfgaclient.ClientWriteRequest{
		Writes: []openfgaclient.ClientTupleKey{
			{
				User:     fmt.Sprintf("user:%s", userID),
				Relation: action,
				Object:   entityID,
			},
		},
	}
	o.logger.Debug("sending FGA create request", "body", body)
	_, err := o.openfgaClient.Write(ctx).
		Body(body).
		Execute()
	if err != nil {
		return errors.Wrap(err, "failed to create relationship in FGA")
	}
	o.logger.Debug("user access created")
	return nil
}

func (o OpenFGAService) ReadUserAccess(ctx context.Context, username string, entityID string, action string) (bool, error) {
	body := openfgaclient.ClientCheckRequest{
		User:     fmt.Sprintf("user:%s", username),
		Relation: action,
		Object:   entityID,
	}
	o.logger.Debug("sending FGA check request", "body", body)
	data, err := o.openfgaClient.Check(ctx).
		Body(body).
		Execute()
	if err != nil {
		return false, errors.Wrap(err, "failed to send check request to FGA")
	}
	o.logger.Debug("received FGA check response", "data", data)
	return *data.Allowed, nil
}

func (o OpenFGAService) ListObjectsUserCanAccess(ctx context.Context, userID string, entityType string, action string) ([]string, error) {
	body := openfgaclient.ClientListObjectsRequest{
		User:     fmt.Sprintf("user:%s", userID),
		Relation: action,
		Type:     entityType,
	}
	o.logger.Debug("sending FGA list objects request", "body", body)
	listObjectsResponse, err := o.openfgaClient.ListObjects(ctx).
		Body(body).
		Execute()
	if err != nil {
		return nil, errors.Wrap(err, "failed to list relationships from FGA")
	}
	o.logger.Debug("user access listed", "objects", listObjectsResponse.Objects)
	return listObjectsResponse.Objects, nil
}

func (o OpenFGAService) DeleteUserAccess(ctx context.Context, userID string, entityID string, action string) error {
	body := openfgaclient.ClientWriteRequest{
		Deletes: []openfgaclient.ClientTupleKeyWithoutCondition{
			{
				User:     fmt.Sprintf("user:%s", userID),
				Relation: action,
				Object:   entityID,
			},
		},
	}
	o.logger.Debug("sending FGA check request", "body", body)
	_, err := o.openfgaClient.Write(ctx).
		Body(body).
		Execute()
	if err != nil {
		return errors.Wrap(err, "failed to delete relationship from FGA")
	}
	o.logger.Debug("user access deleted")
	return nil
}
