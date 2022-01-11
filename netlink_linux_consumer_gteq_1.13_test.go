//go:build linux && go1.13 && !break_circular_dependencies
// +build linux,go1.13,!break_circular_dependencies

package netlink_test

import (
	"errors"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mdlayher/ethtool"
	"github.com/mdlayher/netlink"
	"golang.org/x/sys/unix"
)

func TestIntegrationEthtoolExtendedAcknowledge(t *testing.T) {
	t.Parallel()

	// The ethtool package uses extended acknowledgements and should populate
	// all of netlink.OpError's fields when unwrapped.
	c, err := ethtool.New()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			t.Skip("skipping, ethtool genetlink not available on this system")
		}

		t.Fatalf("failed to open ethtool genetlink: %v", err)
	}

	_, err = c.LinkInfo(ethtool.Interface{Name: "notexist0"})
	if err == nil {
		t.Fatal("expected an error, but none occurred")
	}

	var oerr *netlink.OpError
	if !errors.As(err, &oerr) {
		t.Fatalf("expected wrapped *netlink.OpError, but got: %T", err)
	}

	// Assume the message contents will be relatively static but don't hardcode
	// offset just in case things change.
	if oerr.Offset == 0 {
		t.Fatal("no offset specified in *netlink.OpError")
	}
	oerr.Offset = 0

	want := &netlink.OpError{
		Op:      "receive",
		Err:     unix.ENODEV,
		Message: "no device matches name",
	}

	if diff := cmp.Diff(want, oerr); diff != "" {
		t.Fatalf("unexpected *netlink.OpError (-want +got):\n%s", diff)
	}
}
