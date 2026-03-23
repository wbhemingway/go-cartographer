package cartographer

import (
	"context"
	"net/http"
)

func (c *Client) Render(ctx context.Context, req WorldRequest) (*MapResponse, error) {
	var response MapResponse

	err := c.sendRequest(ctx, http.MethodPost, "/maps", req, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *Client) GetMap(ctx context.Context, mapID string) (*MapResponse, error) {
	var response MapResponse
	endpoint := "/maps/" + mapID

	err := c.sendRequest(ctx, http.MethodGet, endpoint, nil, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *Client) DeleteMap(ctx context.Context, mapID string) error {
	endpoint := "/maps/" + mapID

	return c.sendRequest(ctx, http.MethodDelete, endpoint, nil, nil)
}

func (c *Client) ListMaps(ctx context.Context) ([]MapResponse, error) {
	var responses []MapResponse

	err := c.sendRequest(ctx, http.MethodGet, "/maps", nil, &responses)
	if err != nil {
		return nil, err
	}
	return responses, nil
}
