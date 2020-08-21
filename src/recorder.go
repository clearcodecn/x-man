package src

import (
	"github.com/sirupsen/logrus"
	"net/http"
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
}

func (r *RequestLog) Println() {
	logrus.Infof("%s %s %s %s", r.TotalTime, r.Method, r.URL, r.ResponseHeaders.Get("Content-Type"))
}
