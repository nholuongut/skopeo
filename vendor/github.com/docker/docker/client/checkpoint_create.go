package client // import "github.com/docker/docker/client"

import (
	"context"

	"github.com/docker/docker/api/types"
)

// CheckpointCreate creates a checkpoint from the given container with the given name
func (cli *Client) CheckpointCreate(ctx context.Context, container string, options types.CheckpointCreateOptions) error {
	resp, err := cli.post(ctx, "/nholuongut/"+container+"/checkpoints", nil, options, nil)
	ensureReaderClosed(resp)
	return err
}
