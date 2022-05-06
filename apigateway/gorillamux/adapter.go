package gorillamux

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/awslabs/aws-lambda-go-api-proxy/core"
	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
)

type APIGatewayProxyResponse struct {
	StatusCode        int                 `json:"statusCode"`
	Headers           map[string]string   `json:"headers"`
	MultiValueHeaders map[string][]string `json:"multiValueHeaders"`
	Body              string              `json:"body,omitempty"`
	IsBase64Encoded   bool                `json:"isBase64Encoded,omitempty"`
}

type ErrorHandler interface {
	handleError(err error, statusCode *int) (*APIGatewayProxyResponse, error)
}

type DefaultErrorHandler struct {
	ErrorHandler
}

func (handler *DefaultErrorHandler) handleError(err error, statusCode *int) (*APIGatewayProxyResponse, error) {
	if statusCode == nil {
		response := APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}
		return &response, err
	}

	response := APIGatewayProxyResponse{StatusCode: *statusCode}
	return &response, err
}

type paths struct {
	websocket *string
}

type Adapter struct {
	core.RequestAccessor
	router       *mux.Router
	paths        paths
	errorHandler ErrorHandler
}

func NewAdapter(router *mux.Router) *Adapter {
	adapter := &Adapter{
		router: router,
		paths:  paths{},
	}

	adapter.errorHandler = &DefaultErrorHandler{}

	return adapter
}

func (adapter *Adapter) WithWebsocketPath(path string) *Adapter {
	adapter.paths.websocket = &path
	return adapter
}

func (adapter *Adapter) Handle(ctx context.Context, event map[string]interface{}) (*APIGatewayProxyResponse, error) {
	proxyRequest := events.APIGatewayProxyRequest{}
	err := mapstructure.Decode(event, &proxyRequest)
	if err != nil {
		return adapter.errorHandler.handleError(err, nil)
	}

	if isWebsocketRequest(proxyRequest, event) {
		resp, err := adapter.HandleWebsocketRequest(ctx, proxyRequest, event)
		return resp, err
	}

	resp, err := adapter.HandleRestRequest(ctx, proxyRequest)
	return resp, err
}
