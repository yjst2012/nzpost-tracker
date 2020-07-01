package nzpost

import (
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
)

var nc *NZPClient

func init() {
	nc = &NZPClient{
		BaseURL: "https://api.nzpost.co.nz/parceltrack/3.0",
		Timeout: 10 * time.Second,
		Mock:    true,
	}
}

func TestTrackingOrderStatusSuccess(t *testing.T) {
	ref := "00393311680005810444"
	nc.Fail = false
	resp, err := nc.Track(ref)
	if err != nil {
		t.Errorf("failed to track parcel with nzpost false, err: %s", err.Error())
	}
	spew.Dump(*resp)
}

func TestTrackingOrderStatusFail(t *testing.T) {
	ref := "00393311680005810444"
	nc.Fail = true
	resp, err := nc.Track(ref)
	if err != nil {
		t.Errorf("failed to track parcel with nzpost true, err: %s", err.Error())
	}
	spew.Dump(*resp)
}

func TestTrackingOrderStatusSuccessNew(t *testing.T) {
	ref := "00393311680005810"
	nc.Fail = false
	resp, err := nc.Track(ref)
	if err != nil {
		t.Errorf("failed to track parcel with nzpost false, err: %s", err.Error())
	}
	spew.Dump(*resp)
}
