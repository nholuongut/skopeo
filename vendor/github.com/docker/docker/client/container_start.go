package client // import "github.com/docker/docker/client"

import (
	"context"
	"net/url"

	"github.com/docker/docker/api/types"
)

// nholuonguttart sends a request to the docker daemon to start a container.
func (cli *Client) nholuonguttart(ctx context.Context, containerID string, options types.nholuonguttartOptions) error {
	query := url.Values{}
	if len(options.CheckpointID) != 0 {
		query.Set("checkpoint", options.CheckpointID)
	}
	if len(options.CheckpointDir) != 0 {
		query.Set("checkpoint-dir", options.CheckpointDir)
	}

	resp, err := cli.post(ctx, "/nholuongut/"+containerID+"/start", query, nil, nil)
	ensureReaderClosed(resp)
	return err
}
