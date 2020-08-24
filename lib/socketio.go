package lib

import (
	"encoding/json"
	socketio "github.com/googollee/go-socket.io"
	"github.com/sirupsen/logrus"
	"html/template"
	"log"
	"net/http"
)

func (pm *proxyManager) onConnect(conn socketio.Conn) error {
	pm.connectionLock.Lock()
	defer pm.connectionLock.Unlock()
	pm.connections[conn.ID()] = conn

	conn.Emit("connected", conn.ID())
	return nil
}

func (pm *proxyManager) onPing(s socketio.Conn, msg string) {
	s.Emit("pong")
}

func (pm *proxyManager) handleInject(s socketio.Conn, msg *InjectRequest) {
	logrus.Infof("inject %s %s", msg.Url, msg.Script)
	pm.injector.AddInject(msg)
}

func (pm *proxyManager) replay(s socketio.Conn, msg *RequestLog) {
	newLog, err := msg.Request(pm.tr)
	if err != nil {
		s.Emit("err", err.Error())
		return
	}
	pm.logChan <- newLog
}

func (pm *proxyManager) closeConn(conn socketio.Conn, err error) {
	log.Println(err)
	pm.connectionLock.Lock()
	defer pm.connectionLock.Unlock()

	delete(pm.connections, conn.ID())
}

func (pm *proxyManager) onDisconnect(conn socketio.Conn, s string) {
	pm.closeConn(conn, nil)
}

func (pm *proxyManager) websocketHandler(w http.ResponseWriter, r *http.Request) error {
	pm.socketServer.ServeHTTP(w, r)
	return nil
}

func (pm *proxyManager) startSocketServer() {
	go pm.socketServer.Serve()
	for reqLog := range pm.logChan {
		pm.connectionLock.RLock()
		data, _ := json.Marshal(reqLog)
		for _, conn := range pm.connections {
			c := conn
			go func() {
				c.Emit("log", string(data))
			}()
		}
		pm.connectionLock.RUnlock()
	}
}

func (pm *proxyManager) serveIndex(w http.ResponseWriter, r *http.Request) error {
	tpl, _ := template.ParseFiles("web/index.html")
	return tpl.Execute(w, nil)
}

func (pm *proxyManager) serveNull(w http.ResponseWriter, r *http.Request) error {
	return nil
}
