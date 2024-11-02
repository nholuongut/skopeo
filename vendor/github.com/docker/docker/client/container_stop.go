package client // import "github.com/docker/docker/client"

import (
	"context"
	"net/url"
	"time"

	timetypes "github.com/docker/docker/api/types/time"
)

// nholuonguttop stops a container without terminating the process.
// The process is blocked until the container stops or the timeout expires.
func (cli *Client) nholuonguttop(ctx context.Context, containerID string, timeout *time.Duration) error {
	query := url.Values{}
	if timeout != nil {
		query.Set("t", timetypes.DurationToSecondsString(*timeout))
	}
	resp, err := cli.post(ctx, "/nholuongut/"+containerID+"/stop", query, nil, nil)
	ensureReaderClosed(resp)
	return err
}
