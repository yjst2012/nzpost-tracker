package nzpost

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/viper"
)

const (
	ClientID       = "applywithnzpost" // private license key, need to apply with nzpost
	DefaultBaseURL = "https://api.nzpost.co.nz/parceltrack/3.0"
)

// https://anypoint.mulesoft.com/exchange/portals/nz-post-group/b8271f09-2ad8-4e1c-b6b1-322c5727d148/parceltrack-api/minor/2.0/pages/Track%20a%20Parcel
// error code:
//
// 200001	Partial results returned, not all system(s) have responded
// 200002	All sources responded, data may be incomplete
// 400001	Parameter(s) missing
// 400002	Invalid parameter(s)
// 400003	Non mutually exclusive parameters detected
// 401001	Unauthorised access, please contact administrator
// 500001	General Exception
// 500002	System(s) offline

// NZPClient client
type NZPClient struct {
	BaseURL string
	Timeout time.Duration
	Mock    bool
	Fail    bool
}

// NZPResp Response Parameters
type NZPResp struct {
	Success   bool              `json:"success"`
	Results   NZPResults        `json:"results"`
	Errors    NZPErrorObjParams `json:"errors"`
	MessageID string            `json:"message_id"`
}

// NZPResults Results Response Elements
type NZPResults struct {
	TrackingRef     string              `json:"tracking_reference"`
	MessageID       string              `json:"message_id"`
	MessageDatetime string              `json:"message_datetime"`
	Service         string              `json:"service"`
	Carrier         string              `json:"carrier"`
	TrackingEvents  []NZPTrackingEvents `json:"tracking_events"`
}

// NZPTrackingEvents Tracking Events Response Elements
type NZPTrackingEvents struct {
	TrackingRef       string      `json:"tracking_reference"`
	EventEdifactCode  string      `json:"event_edifact_code"`
	Source            string      `json:"source"`
	Seqref            string      `json:"seqref"`
	StatusDescription string      `json:"status_description"`
	EventDescription  string      `json:"event_description"`
	EventDatetime     string      `json:"event_datetime"`
	SignedBy          NZPSignedBy `json:"signed_by"`
}

// NZPSignedBy signed by
type NZPSignedBy struct {
	Name      string `json:"name"`
	Signature string `json:"signature"`
}

// NZPScanReasonResp Scan Reason Response Elements
type NZPScanReasonResp struct {
	ReasonCode     string `json:"reason_code"`
	ReasonComment1 string `json:"reason_comment1"`
}

// NZPErrorObjParams Error Object Parameters
type NZPErrorObjParams struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details"`
}

// NewClient create new client
func NewClient(baseURL string, timeout time.Duration) NZPClient {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return NZPClient{BaseURL: baseURL, Timeout: timeout}
}

// Track Report sends a message to the Slack channel
func (nc *NZPClient) Track(ref string) (*NZPResp, error) {

	if ref == "" {
		return nil, errors.New("invalid param")
	}
	var respBytes []byte
	var err error
	if !nc.Mock {
		respBytes, err = nc.get("/parcels/" + ref)
		if err != nil {
			return nil, errors.New("Track error")
		}
	} else {
		respBytes, err = nc.getMock("/parcels/"+ref, true, nc.Fail)
		if err != nil {
			return nil, errors.New("TrackMock err")
		}
	}
	resp := &NZPResp{}
	if err := json.Unmarshal(respBytes, resp); err != nil {
		return nil, errors.New("JSON decode error")
	}
	return resp, nil
}

// Get request
func (nc *NZPClient) get(path string) ([]byte, error) {
	httpReq, err := http.NewRequest("GET", nc.BaseURL+path, nil)
	if err != nil {
		log.Printf("nzpost client http.NewRequest GET (%s) ERROR: %v", nc.BaseURL+path, err)
		return nil, errors.New("Error creating GET request")
	}

	//config http client
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("client_id", ClientID)
	httpClient := http.DefaultClient
	if viper.GetBool("allow-insecure-tls-client") {
		httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}
	httpClient.Timeout = nc.Timeout

	//send request
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, errors.New("http response nil response")
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("http response read")
	}
	if respBody == nil {
		return nil, errors.New("http response nil")
	}

	//check if the request is succeeded
	if resp.StatusCode/100 != 2 {
		spew.Dump(*httpReq)
		log.Println(respBody)
		return nil, errors.New("http response err")
	}
	log.Printf("nzpost client GET resp for [path: %s] ----> %s", path, string(respBody))
	return respBody, nil
}

func (nc *NZPClient) getMock(path string, mock, fail bool) ([]byte, error) {

	if !mock {
		return nc.get(path)
	}

	if !fail {
		buf, err := ioutil.ReadFile("./result.success")
		if err != nil {
			log.Panic("failed to open ./result.success")
		}
		return buf, nil
	}
	buf, err := ioutil.ReadFile("./result.fail")
	if err != nil {
		log.Panic("failed to open ./result.fail")
	}
	return buf, nil
}
