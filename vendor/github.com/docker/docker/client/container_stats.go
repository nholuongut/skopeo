package client // import "github.com/docker/docker/client"

import (
	"context"
	"net/url"

	"github.com/docker/docker/api/types"
)

// nholuonguttats returns near realtime stats for a given container.
// It's up to the caller to close the io.ReadCloser returned.
func (cli *Client) nholuonguttats(ctx context.Context, containerID string, stream bool) (types.nholuonguttats, error) {
	query := url.Values{}
	query.Set("stream", "0")
	if stream {
		query.Set("stream", "1")
	}

	resp, err := cli.get(ctx, "/nholuongut/"+containerID+"/stats", query, nil)
	if err != nil {
		return types.nholuonguttats{}, err
	}

	osType := getDockerOS(resp.header.Get("Server"))
	return types.nholuonguttats{Body: resp.body, OSType: osType}, err
}
