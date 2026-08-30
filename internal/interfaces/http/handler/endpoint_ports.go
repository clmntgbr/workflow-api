package handler

import (
	"context"

	endpointcmd "go-api/internal/application/command/endpoint"
	queryendpoint "go-api/internal/application/query/endpoint"
	domainendpoint "go-api/internal/domain/endpoint"
)

type endpointCreateHandler interface {
	Handle(ctx context.Context, cmd endpointcmd.CreateEndpointCommand) (*domainendpoint.Endpoint, error)
}

type endpointImportHandler interface {
	Handle(ctx context.Context, cmd endpointcmd.ImportEndpointsFromOpenAPICommand) ([]domainendpoint.Endpoint, error)
}

type endpointUpdateHandler interface {
	Handle(ctx context.Context, cmd endpointcmd.UpdateEndpointCommand) error
}

type endpointDeleteHandler interface {
	Handle(ctx context.Context, cmd endpointcmd.DeleteEndpointCommand) error
}

type endpointGetByIDHandler interface {
	Handle(ctx context.Context, q queryendpoint.GetEndpointByIDQuery) (*domainendpoint.EndpointView, error)
}

type endpointListByProjectHandler interface {
	Handle(ctx context.Context, q queryendpoint.ListEndpointsByProjectQuery) ([]domainendpoint.EndpointView, int64, error)
}
