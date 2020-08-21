package src

import (
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
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
		filter:      nil,
		certManager: defaultCertManager,
		tr:          tr,
	}
	pm.setup()

	logrus.Info("server listen at: http://127.0.0.1:3344")
	logrus.Info("cert address: http://127.0.0.1:3344/-/cert")
	http.ListenAndServe(":3344", &pm)
}
