// +build nholuongut_image_docker_daemon_stub

package alltransports

import "github.com/nholuongut/image/transports"

func init() {
	transports.Register(transports.NewStubTransport("docker-daemon"))
}
