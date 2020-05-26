package scm

import (
	"context"
	"fmt"

	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/go-scm/scm/factory"
)

// Client represents the go-scm client
type Client struct {
	*scm.Client
}

// NewClient returns an instance of client
func NewClient(url, token string) (*Client, error) {
	driverName, err := getDriverName(url)
	if err != nil {
		return nil, err
	}
	var client *scm.Client
	client, err = factory.NewClient(driverName, "", token)
	if err != nil {
		return nil, err
	}
	return &Client{Client: client}, nil
}

// ListWebhooks returns a list of webhook IDs of the given listener in this repository
func (client *Client) ListWebhooks(name, listenerURL string) ([]string, error) {
	hooks, _, err := client.Repositories.ListHooks(context.Background(), name, scm.ListOptions{})
	if err != nil {
		return nil, err
	}
	ids := []string{}
	for _, hook := range hooks {
		if hook.Target == listenerURL {
			ids = append(ids, hook.ID)
		}
	}
	return ids, nil
}

// DeleteWebhooks deletes all webhooks that associate with the given listener in this repository
func (client *Client) DeleteWebhooks(name string, ids []string) ([]string, error) {
	deleted := []string{}
	for _, id := range ids {
		_, err := client.Repositories.DeleteHook(context.Background(), name, id)
		if err != nil {
			return deleted, fmt.Errorf("failed to delete webhook id %s: %w", id, err)
		}
		deleted = append(deleted, id)
	}
	return deleted, nil
}

// CreateWebhook creates a new webhook in the repository
// It returns ID of the created webhook
func (client *Client) CreateWebhook(name, listenerURL, secret string) (string, error) {
	in := &scm.HookInput{
		Target: listenerURL,
		Secret: secret,
		Events: scm.HookEvents{
			PullRequest: true,
			Push:        true,
		},
	}
	created, _, err := client.Repositories.CreateHook(context.Background(), name, in)
	return created.ID, err
}
