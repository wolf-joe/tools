package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"text/template"
	"time"
)

// HardwareMonitor gen by https://github.com/miku/zek
type HardwareMonitor struct {
	XMLName               xml.Name `xml:"HardwareMonitor"`
	Text                  string   `xml:",chardata"`
	HardwareMonitorHeader struct {
		Text          string `xml:",chardata"`
		Signature     string `xml:"signature"`
		Version       string `xml:"version"`
		HeaderSize    string `xml:"headerSize"`
		EntryCount    string `xml:"entryCount"`
		EntrySize     string `xml:"entrySize"`
		Time          string `xml:"time"`
		GpuEntryCount string `xml:"gpuEntryCount"`
		GpuEntrySize  string `xml:"gpuEntrySize"`
	} `xml:"HardwareMonitorHeader"`
	HardwareMonitorEntries struct {
		Text                 string `xml:",chardata"`
		HardwareMonitorEntry []struct {
			Text              string `xml:",chardata"`
			SrcName           string `xml:"srcName"`
			SrcUnits          string `xml:"srcUnits"`
			LocalizedSrcName  string `xml:"localizedSrcName"`
			LocalizedSrcUnits string `xml:"localizedSrcUnits"`
			RecommendedFormat string `xml:"recommendedFormat"`
			Data              string `xml:"data"`
			MinLimit          string `xml:"minLimit"`
			MaxLimit          string `xml:"maxLimit"`
			Flags             string `xml:"flags"`
			Gpu               string `xml:"gpu"`
			SrcId             string `xml:"srcId"`
		} `xml:"HardwareMonitorEntry"`
	} `xml:"HardwareMonitorEntries"`
	HardwareMonitorGpuEntries struct {
		Text                    string `xml:",chardata"`
		HardwareMonitorGpuEntry struct {
			Text      string `xml:",chardata"`
			GpuId     string `xml:"gpuId"`
			Family    string `xml:"family"`
			Device    string `xml:"device"`
			Driver    string `xml:"driver"`
			BIOS      string `xml:"BIOS"`
			MemAmount string `xml:"memAmount"`
		} `xml:"HardwareMonitorGpuEntry"`
	} `xml:"HardwareMonitorGpuEntries"`
}

const (
	indexContent = `<html>
<head><title>Afterburner Exporter</title></head>
<body>
<h0>Afterburner Exporter</h1>
<p><a href="/metrics">Metrics</a></p>
</body>
</html>`
	metricStub = `# TYPE msi_afterburner_HardwareMonitorEntry gauge
{{range .HardwareMonitorEntry}}msi_afterburner_HardwareMonitorEntry{SrcName="{{.SrcName}}",Gpu="{{.Gpu}}",SrcId="{{.SrcId}}"} {{.Data}}
{{end}}`
)

var (
	httpClient interface {
		Do(*http.Request) (*http.Response, error)
	}
	metricTmpl = template.Must(template.New("test").Parse(metricStub))
	indexData  = []byte(indexContent)
	listen     = flag.String("listen", "0.0.0.0:8090", "Afterburner Exporter listen address")
	target     = flag.String("target", "127.0.0.1:82", "MSI Afterburner Remote Server Address")
	password   = flag.String("password", "17cc95b4017d496f82", "MSI Afterburner Remote Server Password")
)

func init() {
	httpClient = &http.Client{
		Timeout: 500 * time.Millisecond,
	}
}

func indexHandler(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write(indexData)
}

func metricsHandler(w http.ResponseWriter, req *http.Request) {
	retMsg := func(msg string) {
		log.Println(msg)
		_, _ = w.Write([]byte(msg))
	}
	// region forward request
	path := "/mahm"
	req, err := http.NewRequest("GET", "http://"+*target+path, nil)
	if err != nil {
		retMsg(err.Error())
		return
	}
	req.SetBasicAuth("MSIAfterburner", *password)
	var resp *http.Response
	retries := 1
	for retries >= 0 {
		resp, err = httpClient.Do(req)
		if err != nil {
			log.Printf("[%d]get %s failed: %s", retries, path, err)
			retries -= 1
		} else {
			break
		}
	}
	if err != nil {
		retMsg(fmt.Sprintf("get %s failed: %s", path, err))
		return
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		retMsg(fmt.Sprintf("read %s body failed: %s", path, err))
		return
	}
	if resp.StatusCode != http.StatusOK {
		w.WriteHeader(resp.StatusCode)
		_, _ = w.Write(body)
		return
	}
	// endregion

	// region parse content
	data := &HardwareMonitor{}
	if err = xml.Unmarshal(body, data); err != nil {
		retMsg(fmt.Sprintf("unmarshal %s body failed: %s", path, err))
		return
	}
	// endregion

	if err = metricTmpl.Execute(w, data.HardwareMonitorEntries); err != nil {
		retMsg("execute tmpl error: " + err.Error())
		return
	}
	log.Println("send success")
}

func main() {
	flag.Parse()
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/metrics", metricsHandler)
	log.Printf("listen on %s", *listen)
	if err := http.ListenAndServe(*listen, nil); err != nil {
		log.Fatal(err)
	}
}
