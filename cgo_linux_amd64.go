//go:build cgo && linux && amd64

package fdkaac

// #cgo LDFLAGS: -L${SRCDIR}/deps/linux_amd64
// #cgo LDFLAGS: -lfdk-aac -lm
import "C"
