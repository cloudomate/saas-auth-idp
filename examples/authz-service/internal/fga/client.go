package fga

import (
	"context"
	"fmt"

	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
)

// Client wraps the OpenFGA client
type Client struct {
	client  *client.OpenFgaClient
	storeID string
}

// NewClient creates a new OpenFGA client
func NewClient(url, storeID string) (*Client, error) {
	cfg := &client.ClientConfiguration{
		ApiUrl:  url,
		StoreId: storeID,
	}

	fgaClient, err := client.NewSdkClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenFGA client: %w", err)
	}

	return &Client{
		client:  fgaClient,
		storeID: storeID,
	}, nil
}

// Check performs a permission check
// user: "user:<user_id>"
// relation: "can_read", "can_write", "can_manage", etc.
// object: "container:<container_id>" or "document:<doc_id>"
func (c *Client) Check(ctx context.Context, user, relation, object string) (bool, error) {
	body := client.ClientCheckRequest{
		User:     user,
		Relation: relation,
		Object:   object,
	}

	resp, err := c.client.Check(ctx).Body(body).Execute()
	if err != nil {
		return false, fmt.Errorf("permission check failed: %w", err)
	}

	return resp.GetAllowed(), nil
}

// CheckWithTuple performs a permission check with contextual tuples
func (c *Client) CheckWithTuple(ctx context.Context, user, relation, object string, contextualTuples []openfga.TupleKey) (bool, error) {
	body := client.ClientCheckRequest{
		User:     user,
		Relation: relation,
		Object:   object,
	}

	if len(contextualTuples) > 0 {
		body.ContextualTuples = contextualTuples
	}

	resp, err := c.client.Check(ctx).Body(body).Execute()
	if err != nil {
		return false, fmt.Errorf("permission check failed: %w", err)
	}

	return resp.GetAllowed(), nil
}

// WriteTuple writes a relationship tuple
func (c *Client) WriteTuple(ctx context.Context, user, relation, object string) error {
	body := client.ClientWriteRequest{
		Writes: []client.ClientTupleKey{
			{
				User:     user,
				Relation: relation,
				Object:   object,
			},
		},
	}

	_, err := c.client.Write(ctx).Body(body).Execute()
	if err != nil {
		return fmt.Errorf("write tuple failed: %w", err)
	}

	return nil
}

// DeleteTuple deletes a relationship tuple
func (c *Client) DeleteTuple(ctx context.Context, user, relation, object string) error {
	body := client.ClientWriteRequest{
		Deletes: []client.ClientTupleKeyWithoutCondition{
			{
				User:     user,
				Relation: relation,
				Object:   object,
			},
		},
	}

	_, err := c.client.Write(ctx).Body(body).Execute()
	if err != nil {
		return fmt.Errorf("delete tuple failed: %w", err)
	}

	return nil
}

// ListObjects lists objects the user has a specific relation to
func (c *Client) ListObjects(ctx context.Context, user, relation, objectType string) ([]string, error) {
	body := client.ClientListObjectsRequest{
		User:     user,
		Relation: relation,
		Type:     objectType,
	}

	resp, err := c.client.ListObjects(ctx).Body(body).Execute()
	if err != nil {
		return nil, fmt.Errorf("list objects failed: %w", err)
	}

	return resp.GetObjects(), nil
}
