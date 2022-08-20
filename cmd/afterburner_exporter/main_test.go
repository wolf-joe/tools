package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockClient struct {
	body []byte
}

func (c *mockClient) Do(_ *http.Request) (*http.Response, error) {
	return &http.Response{
		Status:           "",
		StatusCode:       http.StatusOK,
		Proto:            "",
		ProtoMajor:       0,
		ProtoMinor:       0,
		Header:           nil,
		Body:             ioutil.NopCloser(bytes.NewReader(c.body)),
		ContentLength:    int64(len(c.body)),
		TransferEncoding: nil,
		Close:            false,
		Uncompressed:     false,
		Trailer:          nil,
		Request:          nil,
		TLS:              nil,
	}, nil
}

func Test_metricsHandler(t *testing.T) {
	httpClient = &mockClient{body: []byte(mockAfterburnerResp)}

	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()
	metricsHandler(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.Fatal(resp.StatusCode)
	}
	expect := `# TYPE msi_afterburner_HardwareMonitorEntry gauge
msi_afterburner_HardwareMonitorEntry{SrcName="Framerate",Gpu="4294967295",SrcId="80"} 0
msi_afterburner_HardwareMonitorEntry{SrcName="GPU1 power",Gpu="0",SrcId="96"} 11
`
	if string(body) != expect {
		t.Fatal(string(body))
	}
}

const mockAfterburnerResp = `
<?xml version="1.0" encoding="utf-8"?>
<HardwareMonitor>
    <HardwareMonitorHeader>
        <signature>1296123981</signature>
        <version>131072</version>
        <headerSize>32</headerSize>
        <entryCount>38</entryCount>
        <entrySize>1324</entrySize>
        <time>1431295640</time>
        <gpuEntryCount>2</gpuEntryCount>
        <gpuEntrySize>1304</gpuEntrySize>
    </HardwareMonitorHeader>
    <HardwareMonitorEntries>
        <HardwareMonitorEntry>
            <srcName>Framerate</srcName>
            <srcUnits>FPS</srcUnits>
            <localizedSrcName>Framerate</localizedSrcName>
            <localizedSrcUnits>FPS</localizedSrcUnits>
            <recommendedFormat>%.0f</recommendedFormat>
            <data>0</data>
            <minLimit>0</minLimit>
            <maxLimit>200</maxLimit>
            <flags>SHOW_IN_OSD</flags>
            <gpu>4294967295</gpu>s
            <srcId>80</srcId>
        </HardwareMonitorEntry>
        <HardwareMonitorEntry>
            <srcName>GPU1 power</srcName>
            <srcUnits>%</srcUnits>
            <localizedSrcName>GPU1 power</localizedSrcName>
            <localizedSrcUnits>%</localizedSrcUnits>
            <recommendedFormat>%.0f</recommendedFormat>
            <data>11</data>
            <minLimit>0</minLimit>
            <maxLimit>150</maxLimit>
            <flags>None</flags>
            <gpu>0</gpu>
            <srcId>96</srcId>
        </HardwareMonitorEntry>
    </HardwareMonitorEntries>
    <HardwareMonitorGpuEntries>
        <HardwareMonitorGpuEntry>
            <gpuId>VEN_10DE&amp;DEV_13C0&amp;SUBSYS_85041043&amp;REV_A1&amp;BUS_1&amp;DEV_0&amp;FN_0</gpuId>
            <family>GM204-A</family>
            <device>GeForce GTX 980</device>
            <driver>350.12</driver>
            <BIOS>84.04.1f.00.02</BIOS>
            <memAmount>0</memAmount>
        </HardwareMonitorGpuEntry>
        <HardwareMonitorGpuEntry>
            <gpuId>VEN_8086&amp;DEV_0162&amp;SUBSYS_01621849&amp;REV_09&amp;BUS_0&amp;DEV_2&amp;FN_0</gpuId>
            <family />
            <device />
            <driver />
            <BIOS />
            <memAmount>0</memAmount>
        </HardwareMonitorGpuEntry>
    </HardwareMonitorGpuEntries>
</HardwareMonitor>
`
