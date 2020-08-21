package src

import (
	"encoding/json"
	"fmt"
	socketio "github.com/googollee/go-socket.io"
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
	fmt.Println("inject", msg.Url, msg.Script)
	pm.injector.AddInject(msg)
}

type ReRequest struct {
	Request *RequestLog `json:"request"`
}

//func (pm *proxyManager) reRequest(s socketio.Conn, msg *ReRequest) {
//	req := http.NewRequest(msg.Request.Method, msg.Request.URL, nil)
//}

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

const mockLog = `{"url":"http://cdn2.clearcode.cn/index.html","requestHeaders":{"Accept-Encoding":["gzip"],"User-Agent":["Go-http-client/1.1"]},"responseHeaders":{"Accept-Ranges":["bytes"],"Access-Control-Allow-Origin":["*"],"Access-Control-Expose-Headers":["X-Log, X-Reqid"],"Access-Control-Max-Age":["2592000"],"Age":["80921"],"Ali-Swift-Global-Savetime":["1597910373"],"Cache-Control":["public, max-age=31536000"],"Connection":["keep-alive"],"Content-Disposition":["inline; filename=\"index.html\"; filename*=utf-8''index.html"],"Content-Encoding":["gzip"],"Content-Length":["1371"],"Content-Md5":["zaG4Hu/qBu0LZhbThuoKaw=="],"Content-Transfer-Encoding":["binary"],"Content-Type":["text/html"],"Date":["Thu, 20 Aug 2020 07:59:33 GMT"],"Eagleid":["7d4a042215979912940197428e"],"Etag":["\"Flhq-OhF8F4FU2OI2qGOUZTrDd9p.gz\""],"Last-Modified":["Tue, 09 Jun 2020 05:29:09 GMT"],"Server":["Tengine"],"Timing-Allow-Origin":["*"],"Vary":["Accept-Encoding"],"Via":["cache34.l2cn2600[93,200-0,M], cache42.l2cn2600[94,0], vcache10.cn679[0,200-0,H], vcache14.cn679[1,0]"],"X-Cache":["HIT TCP_MEM_HIT dirn:11:87175746"],"X-Log":["X-Log"],"X-M-Log":["QNM:xs463;SRCPROXY:xs1752;SRC:39;SRCPROXY:41;QNM3:43"],"X-M-Reqid":["uBkAAFWO99ME6ywW"],"X-Qiniu-Zone":["0"],"X-Qnm-Cache":["Miss"],"X-Reqid":["cbYAAABtEtQE6ywW"],"X-Svr":["IO"],"X-Swift-Cachetime":["2592000"],"X-Swift-Savetime":["Thu, 20 Aug 2020 07:59:33 GMT"]},"createTime":"2020-08-21T14:28:13.997201+08:00","totalTime":33161316,"responseBody":"\u003c!doctype html\u003e\n\u003chtml lang=\"en\"\u003e\n\u003chead\u003e\n    \u003cmeta charset=\"UTF-8\"\u003e\n    \u003cmeta name=\"viewport\"\n          content=\"width=device-width, user-scalable=no, initial-scale=1.0, maximum-scale=1.0, minimum-scale=1.0\"\u003e\n    \u003cmeta http-equiv=\"X-UA-Compatible\" content=\"ie=edge\"\u003e\n    \u003ctitle\u003eDocument\u003c/title\u003e\n    \u003cscript src=\"https://cdn.bootcdn.net/ajax/libs/jquery/3.5.1/jquery.min.js\"\u003e\u003c/script\u003e\n    \u003cstyle\u003e\n        * {\n            margin: 0;\n            padding: 0\n        }\n\n        #app {\n            width: 100%;\n            color: #000;\n            text-align: center;\n            line-height: 100px;\n        }\n\n        .flex-parent {\n            display: flex;\n            margin-bottom: 1px;\n        }\n\n        .flex {\n            flex: 1;\n            width: 33.3333333333%;\n            height: 100px;\n        }\n\n        .flex:not(:first-child) {\n            margin-left: 1px;\n        }\n\n        .chou {\n            background: #FF0066;\n        }\n\n        .zou {\n            background: #FF9900;\n        }\n\n        .ji {\n            background: #99CC33;\n        }\n\n        .gift_b {\n            background: #FFFF66;\n        }\n    \u003c/style\u003e\n\u003c/head\u003e\n\u003cbody\u003e\n\u003cdiv id=\"app\"\u003e\n    \u003cdiv class=\"flex-parent\"\u003e\n        \u003cdiv class=\"flex ji f1\"\u003e\n            吉工家\n        \u003c/div\u003e\n        \u003cdiv class=\"flex zou f2\"\u003e\n            走鸭\n        \u003c/div\u003e\n        \u003cdiv class=\"flex ji f3\"\u003e吉工家\u003c/div\u003e\n    \u003c/div\u003e\n    \u003cdiv class=\"flex-parent\"\u003e\n        \u003cdiv class=\"flex zou f8\"\u003e走鸭\u003c/div\u003e\n        \u003cdiv class=\"flex chou\"\u003e\n            抽奖\n        \u003c/div\u003e\n        \u003cdiv class=\"flex zou f4\"\u003e走鸭\u003c/div\u003e\n    \u003c/div\u003e\n    \u003cdiv class=\"flex-parent\"\u003e\n        \u003cdiv class=\"flex ji f7\"\u003e吉工家\u003c/div\u003e\n        \u003cdiv class=\"flex zou f6\"\u003e走鸭\u003c/div\u003e\n        \u003cdiv class=\"flex ji f5\"\u003e吉工家\u003c/div\u003e\n    \u003c/div\u003e\n    \u003cdiv class=\"flex-parent \" style=\"margin-top: 30px;\"\u003e\n        \u003cdiv class=\"flex jr\"\u003e吉(0)\u003c/div\u003e\n        \u003cdiv class=\"flex zr\"\u003e走(0)\u003c/div\u003e\n    \u003c/div\u003e\n\u003c/div\u003e\n\u003cscript\u003e\n    $(function () {\n        $('.jr').text(\"吉(\" + getJi() + \")\")\n        $('.zr').text(\"走(\" + getZou() + \")\")\n\n        var speed = 150, //跑马灯速度\n            click = true, //阻止多次点击\n            img_index = 1, //阴影停在当前奖品的序号\n            circle = 0, //跑马灯跑了多少次\n            maths; //取一个随机数;\n        $('.chou').on(\"click\", function () {\n            if (click) {\n                click = false;\n                img_index = 1;\n                $('.flex').removeClass('gift_b')\n                maths = parseInt((Math.random() * 10) + 80);\n                light();\n            } else {\n                return false;\n            }\n        });\n\n        function light() {\n            img();\n            circle++;\n            var timer = setTimeout(light, speed);\n            if (circle \u003e 0 \u0026\u0026 circle \u003c 5) {\n                speed -= 10;\n            } else if (circle \u003e 5 \u0026\u0026 circle \u003c 20) {\n                speed -= 5;\n            } else if (circle \u003e 50 \u0026\u0026 circle \u003c 70) {\n                speed += 5\n            } else if (circle \u003e 70 \u0026\u0026 circle \u003c maths) {\n                speed += 10\n            } else if (circle == maths) {\n                var text = $('.f' + img_index).text();\n                clearTimeout(timer);\n                click = true;\n                speed = 150;\n                if (circle % 80 % 2 == 1) {\n                    setJi()\n                    $('.jr').text(\"吉(\" + getJi() + \")\")\n                } else {\n                    setZou()\n                    $('.zr').text(\"走(\" + getZou() + \")\")\n                }\n                circle = 0;\n            }\n        }\n\n        function img() {\n            if (img_index == 9) {\n                $('.f' + (img_index - 1)).removeClass('gift_b')\n                img_index = 1;\n            }\n            $('.f' + img_index).addClass('gift_b')\n            if (img_index \u003e 1) {\n                $('.f' + (img_index - 1)).removeClass('gift_b')\n            }\n            img_index++;\n        }\n\n        function setJi() {\n            ji = getJi()\n            ji++\n            localStorage.setItem(\"ji\", ji)\n        }\n\n        function setZou() {\n            ji = getZou()\n            ji++\n            localStorage.setItem(\"zou\", ji)\n        }\n\n        function getJi() {\n            return localStorage.getItem(\"ji\") || 0\n        }\n\n        function getZou() {\n            return localStorage.getItem(\"zou\") || 0\n        }\n    });\n\n\u003c/script\u003e\n\u003c/body\u003e\n\u003c/html\u003e","method":"GET","status":200}`

func (pm *proxyManager) startSocketServer() {
	go pm.socketServer.Serve()
	//var mockFunc = func() {
	//	for {
	//		time.Sleep(5 * time.Second)
	//		pm.connectionLock.RLock()
	//		for _, conn := range pm.connections {
	//			c := conn
	//			go func() {
	//				c.Emit("log", mockLog)
	//			}()
	//		}
	//		pm.connectionLock.RUnlock()
	//	}
	//}
	//if true {
	//	mockFunc()
	//}
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
