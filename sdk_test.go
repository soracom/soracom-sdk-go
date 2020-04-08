package soracom

import (
	"bytes"
	"testing"
)

func TestMarshals(t *testing.T) {
	t.Run("ListSessionEvent", func(t *testing.T) {

		testdata := `[
  {
    "imsi": "001050910800000",
    "time": 1574853374100,
    "createdTime": "2019-11-27T07:15:19.544Z",
    "operatorId": "OP9999209600",
    "event": "Deleted",
    "ueIpAddress": "10.154.128.64",
    "imei": "999916091541232",
    "apn": "soracom.io",
    "dns0": "100.127.0.53",
    "dns1": "100.127.1.53",
    "cell": {
      "radioType": "lte",
      "mcc": 440,
      "mnc": 10,
      "tac": 21,
      "eci": 67387471
    },
    "primaryImsi": "001050910800000"
  },
  {
    "imsi": "001050910800000",
    "time": 1574838919544,
    "createdTime": "2019-11-27T07:15:19.544Z",
    "operatorId": "OP9999209600",
    "event": "Created",
    "ueIpAddress": "10.154.128.64",
    "imei": "999916091541232",
    "apn": "soracom.io",
    "dns0": "100.127.0.53",
    "dns1": "100.127.1.53",
    "cell": {
      "radioType": "lte",
      "mcc": 440,
      "mnc": 10,
      "tac": 21,
      "eci": 67387471
    },
    "primaryImsi": "001050910800000"
  },
  {
    "imsi": "001050910800000",
    "time": 1574838918340,
    "createdTime": "2019-11-27T07:14:17.717Z",
    "operatorId": "OP9999209600",
    "event": "Deleted",
    "ueIpAddress": "10.154.128.64",
    "imei": "999916091541232",
    "apn": "soracom.io",
    "dns0": "100.127.0.53",
    "dns1": "100.127.1.53",
    "cell": {
      "radioType": "lte",
      "mcc": 440,
      "mnc": 10,
      "tac": 21,
      "eci": 67387471
    },
    "primaryImsi": "001050910800000"
  },
  {
    "imsi": "001050910800000",
    "time": 1574838857717,
    "createdTime": "2019-11-27T07:14:17.717Z",
    "operatorId": "OP9999209600",
    "event": "Created",
    "ueIpAddress": "10.154.128.64",
    "imei": "999916091541232",
    "apn": "soracom.io",
    "dns0": "100.127.0.53",
    "dns1": "100.127.1.53",
    "cell": {
      "radioType": "lte",
      "mcc": 440,
      "mnc": 10,
      "tac": 21,
      "eci": 67387471
    },
    "primaryImsi": "001050910800000"
  }
]`

		b := bytes.NewBufferString(testdata)
		rs, err := parseListSessionEvents(b)
		if err != nil {
			t.Fatalf("failed to parseListSessionEvents(): %s", err)
		}
		if len(rs) != 4 {
			t.Fatalf("result length: want 4 got %d", len(rs))
		}
	})
}
