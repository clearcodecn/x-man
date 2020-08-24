package lib

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

type InjectRequest struct {
	Method  string `json:"method"`
	Url     string `json:"url"`
	Script  string `json:"script"`
	Options struct {
		IgnoreQuery    bool `json:"ignoreQuery"`
		RememberInject bool `json:"rememberInject"`
	} `json:"options"`
}

type Injector interface {
	AddInject(request *InjectRequest)
	Inject(req *http.Request, response *http.Response) (*http.Response, bool, error)
}

type javascriptInjector struct {
	urls map[string]*InjectRequest // key = method+url,
	sync.RWMutex
}

func (j *javascriptInjector) Inject(req *http.Request, response *http.Response) (*http.Response, bool, error) {
	// 1. try full url.
	key := fmt.Sprintf("%s%s", req.Method, strings.TrimSuffix(req.URL.String(), "/"))
	j.RLock()
	iq, ok := j.urls[key]
	j.RUnlock()
	if !ok {
		// 2. try ignore query
		u := &url.URL{}
		*u = *req.URL
		u.RawQuery = ""
		u.Fragment = ""
		key = fmt.Sprintf("%s%s", req.Method, u.String())

		j.RLock()
		iq, ok = j.urls[key]
		j.RUnlock()
		if !ok {
			return response, false, nil
		}
	}
	if !iq.Options.RememberInject {
		j.Lock()
		delete(j.urls, key)
		j.Unlock()
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, false, err
	}
	buf := bytes.NewBuffer(data)
	buf.WriteString(iq.Script)
	response.ContentLength = int64(buf.Len())
	response.Body = ioutil.NopCloser(buf)
	return response, true, nil
}

func (j *javascriptInjector) AddInject(req *InjectRequest) {
	j.Lock()
	defer j.Unlock()
	if req.Options.IgnoreQuery {
		u, err := url.Parse(req.Url)
		if err != nil {
			logrus.Errorf("parse url failed: ", err)
			return
		}
		u.RawQuery = ""
		u.Fragment = ""
		req.Url = strings.TrimSuffix(u.String(), "/")
	}
	j.urls[fmt.Sprintf("%s%s", req.Method, req.Url)] = req
}

func newJavascriptInjector() Injector {
	ji := new(javascriptInjector)
	ji.urls = make(map[string]*InjectRequest)
	return ji
}
