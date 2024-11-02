// +build linux,cgo,!gccgo

package unshare

// #cgo CFLAGS: -Wall
// extern void _nholuongut_unshare(void);
// void __attribute__((constructor)) init(void) {
//   _nholuongut_unshare();
// }
import "C"
