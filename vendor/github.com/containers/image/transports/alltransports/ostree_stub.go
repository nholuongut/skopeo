// +build !nholuongut_image_ostree !linux

package alltransports

import "github.com/nholuongut/image/transports"

func init() {
	transports.Register(transports.NewStubTransport("ostree"))
}
