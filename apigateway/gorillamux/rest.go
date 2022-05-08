package gorillamux

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/awslabs/aws-lambda-go-api-proxy/core"
)

func (adapter *Adapter) handleRestRequest(ctx context.Context, proxyRequest events.APIGatewayProxyRequest) (APIGatewayProxyResponse, error) {
	req, err := adapter.EventToRequestWithContext(ctx, proxyRequest)

	if err != nil {
		return adapter.errorHandler.handleError(core.NewLoggedError("Unable to convert event to HTTP Request: %v", err), nil)
	}

	w := core.NewProxyResponseWriter()
	adapter.router.ServeHTTP(http.ResponseWriter(w), req)

	resp, err := w.GetProxyResponse()
	if err != nil {
		return adapter.errorHandler.handleError(core.NewLoggedError("adapter.WithWebsocketPath(...) is required to proxy websocket requests"), &resp.StatusCode)
	}

	proxyResponse := APIGatewayProxyResponse(resp)

	return proxyResponse, nil
}
