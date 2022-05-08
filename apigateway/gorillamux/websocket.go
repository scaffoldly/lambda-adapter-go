package gorillamux

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/awslabs/aws-lambda-go-api-proxy/core"
	"github.com/mitchellh/mapstructure"
)

func isWebsocketRequest(proxyRequest events.APIGatewayProxyRequest, event map[string]interface{}) bool {
	if _, ok := proxyRequest.Headers["Sec-WebSocket-Version"]; ok {
		return true
	}

	if _, ok := event["requestContext"].(map[string]interface{})["connectionId"]; ok {
		return true
	}

	return false
}

func (adapter *Adapter) handleWebsocketRequest(ctx context.Context, proxyRequest events.APIGatewayProxyRequest, event map[string]interface{}) (APIGatewayProxyResponse, error) {
	if adapter.paths.websocket == nil {
		return adapter.errorHandler.handleError(core.NewLoggedError("adapter.WithWebsocketPath(...) is required to proxy websocket requests"), nil)
	}

	websocketEvt := events.APIGatewayWebsocketProxyRequest{}
	err := mapstructure.Decode(event, &websocketEvt)

	if err != nil {
		return adapter.errorHandler.handleError(core.NewLoggedError("adapter.WithWebsocketPath(...) is required to proxy websocket requests"), nil)
	}

	proxyRequest.Path = *adapter.paths.websocket

	if proxyRequest.Headers == nil {
		proxyRequest.Headers = map[string]string{}
		proxyRequest.MultiValueHeaders = map[string][]string{}
	}
	proxyRequest.Headers["X-ConnectionId"] = websocketEvt.RequestContext.ConnectionID
	proxyRequest.MultiValueHeaders["X-ConnectionId"] = []string{websocketEvt.RequestContext.ConnectionID}

	if websocketEvt.RequestContext.RouteKey == "$connect" {
		proxyRequest.HTTPMethod = http.MethodPut
	} else if websocketEvt.RequestContext.RouteKey == "$disconnect" {
		proxyRequest.HTTPMethod = http.MethodDelete
	} else {
		proxyRequest.HTTPMethod = http.MethodPost
	}

	return adapter.handleRestRequest(ctx, proxyRequest)
}
