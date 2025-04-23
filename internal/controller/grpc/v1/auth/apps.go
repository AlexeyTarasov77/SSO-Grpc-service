package auth

import (
	"context"

	ssov1 "github.com/AlexeySHA256/protos/gen/go/sso"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sso.service/internal/entity"
	"sso.service/pkg/validator"
)

func (s *AuthServer) GetOrCreateApp(ctx context.Context, req *ssov1.GetOrCreateAppRequest) (*ssov1.GetOrCreateAppResponse, error) {
	validationRules := map[string]string{
		"Name":        "required,max=70",
		"Description": "required,max=300",
		"Secret":      "required,min=12,max=64",
	}
	if errs := validator.Validate(req, validationRules); errs != validator.EmptyErrors {
		return nil, status.Error(codes.InvalidArgument, errs)
	}

	data, err := s.service.GetOrCreateApp(ctx, &entity.App{
		Name:        req.GetName(),
		Description: req.GetDescription(),
		Secret:      req.GetSecret(),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get or create app")
	}

	return &ssov1.GetOrCreateAppResponse{
		Id:      data.AppID,
		Created: data.IsCreated,
	}, nil
}
