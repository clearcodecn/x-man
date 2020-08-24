package lib

import (
	"bufio"
	"bytes"
	"compress/flate"
	"compress/gzip"
	"crypto/tls"
	"github.com/andybalholm/brotli"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"
)

var tunnelEstablishedResponseLine = []byte("HTTP/1.1 200 Connection established\r\n\r\n")

const (
	maxRecordSize = 2 * 1024 * 1024 // 2MB
)

type handleFunc func(w http.ResponseWriter, r *http.Request) error

func (pm *proxyManager) serveCert(w http.ResponseWriter, r *http.Request) error {
	w.Header().Add("Content-Type", "application/x-x509-ca-cert")
	w.Header().Add("Content-Disposition", `attachment; filename="cert.pem"`)
	_, err := w.Write(pm.certManager.RootRaw())
	return err
}

func (pm *proxyManager) isSelf(r *http.Request) bool {
	// TODO, pass self host:port
	return false
}

func (pm *proxyManager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logrus.Infof("%s  %s", r.Method, r.URL.Path)
	if h, ok := pm.internalApi[r.URL.Path]; ok {
		pm.serveError(w, r, h)
		return
	}
	if pm.isSelf(r) {
		return
	}
	switch r.Method {
	case http.MethodConnect:
		pm.serveError(w, r, pm.handleHttps)
	default:
		pm.serveError(w, r, pm.handleHttp)
	}
}

func (pm *proxyManager) handleHttps(w http.ResponseWriter, r *http.Request) error {
	conf, err := pm.certManager.GenerateTlsByHost(r.URL.Host)
	if err != nil {
		return err
	}
	httpConn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		return nil
	}
	_, err = httpConn.Write(tunnelEstablishedResponseLine)
	if err != nil {
		return err
	}

	conn := tls.Server(httpConn, conf)
	if err := conn.Handshake(); err != nil {
		return err
	}

	request, err := http.ReadRequest(bufio.NewReader(conn))
	if err != nil {
		return err
	}

	request.RequestURI = ""
	request.URL.Host = r.URL.Host
	request.URL.Scheme = `https`

	err = pm.handleRequest(conn, request)
	conn.Close()
	return err
}

func (pm *proxyManager) handleHttp(w http.ResponseWriter, r *http.Request) error {
	httpConn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		return err
	}
	return pm.handleRequest(httpConn, r)
}

func (pm *proxyManager) copyResponse(response *http.Response) (*http.Response, bool, error) {
	res := &http.Response{
		Proto:            "HTTP/1.1",
		ProtoMajor:       1,
		ProtoMinor:       1,
		StatusCode:       response.StatusCode,
		ContentLength:    response.ContentLength,
		TransferEncoding: response.TransferEncoding,
	}
	isStreamMode := false
	// 1. 看contentLength 是否为空.
	hasContentLength := response.ContentLength != -1
	if hasContentLength {
		isStreamMode = response.ContentLength > maxRecordSize
	}

	if hasContentLength {
		if isStreamMode {
			// do nothing.
			return response, isStreamMode, nil
		} else {
			r, err := getReader(response.Header.Get("Content-Encoding"), response.Body)
			if err != nil {
				return nil, false, err
			}
			data, err := ioutil.ReadAll(r)
			if err != nil {
				return nil, false, err
			}
			res.Header = copyHeader(response.Header)
			res.ContentLength = int64(len(data))
			res.Body = ioutil.NopCloser(bytes.NewReader(data))
			return res, false, nil
		}
	} else {
		// gzip,or unknown mode.
		var buf = bytes.NewBuffer(nil)
		for {
			var b = make([]byte, 2048)
			n, err := response.Body.Read(b)
			if err == io.EOF {
				buf.Write(b[:n])
				break
			}
			buf.Write(b[:n])
			if buf.Len() > maxRecordSize {
				isStreamMode = true
				break
			}
		}
		if isStreamMode {
			r := newReadCloser(buf, response.Body)
			res.Header = copyHeader(response.Header)
			res.Body = r
		} else {
			r, err := getReader(response.Header.Get("Content-Encoding"), buf)
			if err != nil {
				return nil, false, err
			}
			data, err := ioutil.ReadAll(r)
			if err != nil {
				return nil, false, err
			}
			buf.Reset()
			buf.Write(data)
			res.Header = copyHeader(response.Header)
			res.ContentLength = int64(buf.Len())
			res.Body = ioutil.NopCloser(buf)
		}
	}
	return res, isStreamMode, nil
}

func (pm *proxyManager) handleRequest(conn net.Conn, r *http.Request) error {
	var (
		reqLog   *RequestLog
		response *http.Response
	)
	reqLog = &RequestLog{
		URL:            r.URL.String(),
		RequestHeaders: copyHeader(r.Header),
		CreateTime:     time.Now(),
		Method:         r.Method,
	}

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	reqLog.RequestBody = string(reqBody)
	r.Body = ioutil.NopCloser(bytes.NewReader(reqBody))

	newReq := copyRequest(r)
	response, err = pm.tr.RoundTrip(newReq)
	if err != nil {
		return err
	}
	resp, isStreamMode, err := pm.copyResponse(response)
	if err != nil {
		return err
	}
	if isStreamMode {
		reqLog.ResponseBody = "(too big)"
	} else {
		resp, ok, err := pm.injector.Inject(r, resp)
		if err != nil {
			return err
		}
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		resp.Body = ioutil.NopCloser(bytes.NewBuffer(data))
		reqLog.ResponseBody = string(data)
		reqLog.Injected = ok
		resp.Header.Del("Content-Encoding")
	}

	resp.Write(conn)
	reqLog.TotalTime = time.Now().Sub(reqLog.CreateTime)
	reqLog.Status = resp.StatusCode
	reqLog.ResponseHeaders = copyHeader(resp.Header)
	reqLog.Println()

	pm.logChan <- reqLog
	return nil
}

func getReader(encoding string, r io.Reader) (io.ReadCloser, error) {
	var rc io.ReadCloser
	switch encoding {
	case "gzip":
		rd, err := gzip.NewReader(r)
		if err != nil {
			return nil, err
		}
		rc = ioutil.NopCloser(rd)
	case "deflate":
		rc = flate.NewReader(r)
	case "br":
		rd := brotli.NewReader(r)
		rc = ioutil.NopCloser(rd)
	default:
		rc = ioutil.NopCloser(r)
	}
	return rc, nil
}

func copyRequest(req *http.Request) *http.Request {
	req2 := new(http.Request)
	*req2 = *req
	req2.URL = new(url.URL)
	*req2.URL = *req.URL
	req2.Header = make(http.Header, len(req.Header))
	for k, s := range req.Header {
		req2.Header[k] = append([]string(nil), s...)
	}
	return req2
}

func copyHeader(h http.Header) http.Header {
	var newHeader = make(http.Header)
	for k, v := range h {
		newHeader[k] = append([]string(nil), v...)
	}
	return newHeader
}