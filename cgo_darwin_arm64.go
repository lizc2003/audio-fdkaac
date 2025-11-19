//go:build cgo && darwin && arm64

package fdkaac

// #cgo LDFLAGS: -L${SRCDIR}/deps/darwin_arm64
// #cgo LDFLAGS: -lfdk-aac
import "C"
