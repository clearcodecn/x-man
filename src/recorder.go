package src

import (
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"time"
)

type RequestLog struct {
	URL             *url.URL      `json:"url"`
	RequestHeaders  http.Header   `json:"requestHeaders"`
	ResponseHeaders http.Header   `json:"responseHeaders"`
	CreateTime      time.Time     `json:"createTime"`
	TotalTime       time.Duration `json:"totalTime"`
	ResponseBody    string        `json:"responseBody"`
	Method          string        `json:"method"`
}

func (r *RequestLog) Println() {
	logrus.Infof("%s %s %s %s", r.TotalTime, r.Method, r.URL, r.ResponseHeaders.Get("Content-Type"))
}
