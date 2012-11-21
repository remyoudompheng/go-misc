// weechat/web implements a web interface for weechat.
package web

import (
      "net/http"
      "flag"

      ws "code.google.com/p/go.net/websocket"
)

func Register(mux *http.ServeMux) {
      mux.HandleFunc("/weechat", handleHome)
      mux.HandleFunc("/weechat/json", handleJson)
      mux.Handle("/weechat/ws", ws.Handler(handleWebsocket))
}

var (
      weechatAddr string
)

func init() {
      flag.StringVar(&weechatAddr, "weechat.relay", "", "address of Weechat relay")
}

func handleHome(w http.ResponseWriter, req *http.Request) {
}

func handleJson(w http.ResponseWriter, req *http.Request) {
}

func handleWebsocket(conn *ws.Conn) {
}
