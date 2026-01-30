package authz

import (
	"context"
	"fmt"

	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
)

// OpenFGAClient wraps the OpenFGA client for authorization checks
type OpenFGAClient struct {
	client  *client.OpenFgaClient
	storeID string
}

// NewOpenFGAClient creates a new OpenFGA client
func NewOpenFGAClient(url, storeID string) (*OpenFGAClient, error) {
	if storeID == "" {
		return nil, fmt.Errorf("store ID is required")
	}

	cfg := &client.ClientConfiguration{
		ApiUrl:  url,
		StoreId: storeID,
	}

	fgaClient, err := client.NewSdkClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenFGA client: %w", err)
	}

	return &OpenFGAClient{
		client:  fgaClient,
		storeID: storeID,
	}, nil
}

// Check performs a permission check
// Example: Check("user:123", "can_read", "document:doc-1")
func (c *OpenFGAClient) Check(user, relation, object string) (bool, error) {
	body := client.ClientCheckRequest{
		User:     user,
		Relation: relation,
		Object:   object,
	}

	response, err := c.client.Check(context.Background()).Body(body).Execute()
	if err != nil {
		return false, fmt.Errorf("check failed: %w", err)
	}

	return response.GetAllowed(), nil
}

// WriteTuple writes a relationship tuple to OpenFGA
func (c *OpenFGAClient) WriteTuple(user, relation, object string) error {
	body := client.ClientWriteRequest{
		Writes: []client.ClientTupleKey{
			{
				User:     user,
				Relation: relation,
				Object:   object,
			},
		},
	}

	_, err := c.client.Write(context.Background()).Body(body).Execute()
	if err != nil {
		return fmt.Errorf("write failed: %w", err)
	}

	return nil
}

// DeleteTuple deletes a relationship tuple from OpenFGA
func (c *OpenFGAClient) DeleteTuple(user, relation, object string) error {
	body := client.ClientWriteRequest{
		Deletes: []client.ClientTupleKeyWithoutCondition{
			{
				User:     user,
				Relation: relation,
				Object:   object,
			},
		},
	}

	_, err := c.client.Write(context.Background()).Body(body).Execute()
	if err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}

	return nil
}

// ListObjects lists objects of a given type that a user has access to
func (c *OpenFGAClient) ListObjects(user, relation, objectType string) ([]string, error) {
	body := client.ClientListObjectsRequest{
		User:     user,
		Relation: relation,
		Type:     objectType,
	}

	response, err := c.client.ListObjects(context.Background()).Body(body).Execute()
	if err != nil {
		return nil, fmt.Errorf("list objects failed: %w", err)
	}

	return response.GetObjects(), nil
}

// ListRelations lists relations a user has on an object
func (c *OpenFGAClient) ListRelations(user, object string, relations []string) (map[string]bool, error) {
	result := make(map[string]bool)

	for _, relation := range relations {
		allowed, err := c.Check(user, relation, object)
		if err != nil {
			return nil, err
		}
		result[relation] = allowed
	}

	return result, nil
}

// Expand gets the users/usersets that have a relationship with an object
func (c *OpenFGAClient) Expand(relation, object string) (*openfga.UsersetTree, error) {
	body := client.ClientExpandRequest{
		Relation: relation,
		Object:   object,
	}

	response, err := c.client.Expand(context.Background()).Body(body).Execute()
	if err != nil {
		return nil, fmt.Errorf("expand failed: %w", err)
	}

	return response.Tree, nil
}
