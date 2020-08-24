package src

import (
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type RequestLog struct {
	URL             string        `json:"url"`
	RequestHeaders  http.Header   `json:"requestHeaders"`
	ResponseHeaders http.Header   `json:"responseHeaders"`
	CreateTime      time.Time     `json:"createTime"`
	TotalTime       time.Duration `json:"totalTime"`
	RequestBody     string        `json:"requestBody"`
	ResponseBody    string        `json:"responseBody"`
	Method          string        `json:"method"`
	Status          int           `json:"status"`
	Injected        bool          `json:"injected"`
	Replay          bool          `json:"replay"`
}

func (r *RequestLog) Println() {
	logrus.Infof("%s %s %s %s", r.TotalTime, r.Method, r.URL, r.ResponseHeaders.Get("Content-Type"))
}

func (r *RequestLog) Request(rt http.RoundTripper) (*RequestLog, error) {
	newLog := new(RequestLog)
	*newLog = *r
	newLog.CreateTime = time.Now()

	body := strings.NewReader(r.RequestBody)
	req, err := http.NewRequest(r.Method, r.URL, body)
	if err != nil {
		return nil, err
	}
	req.Header = r.RequestHeaders
	resp, err := rt.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	resp.Body, err = getReader(resp.Header.Get("Content-Encoding"), resp.Body)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	newLog.ResponseHeaders = resp.Header
	newLog.ResponseBody = string(data)
	newLog.Status = resp.StatusCode
	newLog.TotalTime = time.Now().Sub(newLog.CreateTime)
	newLog.Replay = true

	return newLog, nil
}
