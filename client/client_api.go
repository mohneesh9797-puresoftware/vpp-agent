package client

import (
	"context"

	"github.com/ligato/vpp-agent/api/models"
)

// SyncClient defines the client-side interface for sync service.
type SyncClient interface {
	// ListCapabilities retrieves supported capabilities.
	ListCapabilities() (map[string][]models.Model, error)

	// ResyncRequest returns new request used for resync.
	ResyncRequest() ResyncRequest

	// ChangeRequest returns new request used for changes.
	ChangeRequest() ChangeRequest
}

// SyncRequest defines common sync request interface.
type SyncRequest interface {
	// Send sends the request.
	Send(ctx context.Context) error
}

// ResyncRequest defines interface for a request used for resync.
type ResyncRequest interface {
	SyncRequest

	// Put puts given models to the request.
	Put(models ...models.ProtoItem)
}

// ChangeRequest defines interface for a request used for resync.
type ChangeRequest interface {
	SyncRequest

	// Update updates given models in the request.
	Update(models ...models.ProtoItem)

	// Delete deletes given models from the request.
	Delete(models ...models.ProtoItem)
}
