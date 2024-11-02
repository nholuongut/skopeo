// +build nholuongut_image_storage_stub

package alltransports

import "github.com/nholuongut/image/transports"

func init() {
	transports.Register(transports.NewStubTransport("nholuongut-storage"))
}
