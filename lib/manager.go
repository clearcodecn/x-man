package lib

import (
	"bufio"
	"bytes"
	"fmt"
	socketio "github.com/googollee/go-socket.io"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"net/http"
	"sync"
	"time"
)

func Main(args []string) {
	var tr http.RoundTripper = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	pm := proxyManager{
		certManager: defaultCertManager,
		tr:          tr,
	}

	pm.socketServer, _ = socketio.NewServer(nil)
	pm.socketServer.OnConnect("", pm.onConnect)
	pm.socketServer.OnEvent("", "ping", pm.onPing)
	pm.socketServer.OnDisconnect("", pm.onDisconnect)
	pm.socketServer.OnError("", pm.closeConn)
	pm.socketServer.OnEvent("", "inject", pm.handleInject)
	pm.socketServer.OnEvent("", "replay", pm.replay)

	pm.logChan = make(chan *RequestLog, 1024)
	pm.connections = make(map[string]socketio.Conn)

	pm.internalApi = make(map[string]handleFunc)
	pm.internalApi["/-/cert"] = pm.serveCert
	pm.internalApi["/-/socket.io"] = pm.websocketHandler
	pm.internalApi["/-/socket.io/"] = pm.websocketHandler
	pm.internalApi["/-/"] = pm.serveIndex
	pm.internalApi["/favicon.ico"] = pm.serveNull
	pm.injector = newJavascriptInjector()

	go pm.startSocketServer()

	logrus.Info("server listen at: http://127.0.0.1:3344")
	logrus.Info("cert address: http://127.0.0.1:3344/-/cert")
	logrus.Info("ui page: http://127.0.0.1:3344/-/")
	http.ListenAndServe(":3344", &pm)
}

type proxyManager struct {
	certManager *certManager

	tr           http.RoundTripper
	internalApi  map[string]handleFunc
	socketServer *socketio.Server

	connectionLock sync.RWMutex
	connections    map[string]socketio.Conn
	logChan        chan *RequestLog

	injector Injector
}

type writer struct {
	http.ResponseWriter
	isWrite bool
}

func (w *writer) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.isWrite = true
}

func (w *writer) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}

func (pm *proxyManager) serveError(w http.ResponseWriter, r *http.Request, h handleFunc) {
	localWriter := &writer{ResponseWriter: w}
	if err := h(localWriter, r); err != nil {
		if !localWriter.isWrite {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("proxy failed: %s", err)))
		}
	}
	return
}

type readCloser struct {
	buf  *bytes.Buffer
	body io.ReadCloser
}

func (r *readCloser) Read(b []byte) (int, error) {
	if r.buf.Len() != 0 {
		return r.buf.Read(b)
	}
	return r.body.Read(b)
}

func (r *readCloser) Close() error {
	return r.body.Close()
}

func newReadCloser(buf *bytes.Buffer, body io.ReadCloser) io.ReadCloser {
	return &readCloser{buf: buf, body: body}
}
