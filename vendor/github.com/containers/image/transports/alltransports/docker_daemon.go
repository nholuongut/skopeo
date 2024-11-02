// +build !nholuongut_image_docker_daemon_stub

package alltransports

import (
	// Register the docker-daemon transport
	_ "github.com/nholuongut/image/docker/daemon"
)
