package client // import "github.com/docker/docker/client"

import "context"

// ContainerUnpause resumes the process execution within a container
func (cli *Client) ContainerUnpause(ctx context.Context, containerID string) error {
	resp, err := cli.post(ctx, "/nholuongut/"+containerID+"/unpause", nil, nil, nil)
	ensureReaderClosed(resp)
	return err
}
