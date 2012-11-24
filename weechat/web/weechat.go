// weechat/web implements a web interface for weechat.
package web

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"sync"

	ws "code.google.com/p/go.net/websocket"
	"github.com/remyoudompheng/go-misc/weechat"
)

func Register(mux *http.ServeMux) {
	if mux == nil {
		mux = http.DefaultServeMux
	}
	mux.HandleFunc("/weechat", handleHome)
	mux.HandleFunc("/weechat/buflines", handleLines)
	mux.Handle("/weechat/ws", ws.Handler(handleWebsocket))
}

var (
	weechatAddr string
	weechatConn *weechat.Conn
	initOnce    sync.Once
)

func init() {
	flag.StringVar(&weechatAddr, "weechat.relay", "", "address of Weechat relay")
}

func initWeechat() {
	conn, err := weechat.Dial(weechatAddr)
	if err != nil {
		log.Fatal(err)
	}
	weechatConn = conn
}

const homeTplStr = `
<!DOCTYPE html>
<html>
  <head>
    <title>Weechat</title>
    <link href="/libs/bootstrap/css/bootstrap.min.css" rel="stylesheet">
    <script src="/libs/jquery.min.js"></script>
    <script src="/libs/bootstrap/js/bootstrap.min.js"></script>
    <script type="text/javascript">
    $(document).ready(function() {
      $("ul#buffers li").click(function() {
            $.get("/weechat/buflines",
            {"buffer": $(this).attr("addr")},
            function(data) {
                  $("ul#lines").html(data);
            });
      });
    });
    </script>
  </head>
  <body style="padding-top: 60px;">
    <div class="navbar navbar-inverse navbar-fixed-top">
      <div class="navbar-inner">
      <div class="container-fluid">
      <div class="nav-collapse collapse">
        <ul class="nav">
          <li><a href="/">Home</a></li>
          <li class="active"><a href="#">Weechat</a></li>
        </ul>
      </div>
      </div>
      </div>
    </div>

    <div class="container-fluid">
    <div class="row-fluid">
      <div class="span3"><!-- left, vertical -->
      <div class="sidebar-nav well">
        <ul class="nav nav-list" id="buffers">
          {{ range $buf := $ }}
          <li addr="{{ $buf.Self | printf "%x" }}"><a href="javascript:void(0);">{{ $buf.Name }}</a></li>
          {{ end }}
        </ul>
      </div>
      </div>

      <div class="span9">
        <div class="span12"><!-- title -->
          <h1>Weechat</h1>
        </div>

        <div class="span12">
          <!-- buffer lines -->
          <ul id="lines"></ul>
        </div>
      </div>
    </div>
    </div>
  </body>
</html>
`

var homeTpl = template.Must(template.New("home").Parse(homeTplStr))

func handleHome(w http.ResponseWriter, req *http.Request) {
	initOnce.Do(initWeechat)
	bufs, err := weechatConn.ListBuffers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	homeTpl.Execute(w, bufs)
}

const linesTplStr = `
{{ range $line := $ }}
<li>{{ $line.Date.Format "2006-01-02 15:04:05" }} {{ $line.Message }}</li>
{{ end }}
`

var linesTpl = template.Must(template.New("lines").Parse(linesTplStr))

func handleLines(w http.ResponseWriter, req *http.Request) {
	initOnce.Do(initWeechat)
	req.ParseForm()
	bufHex := req.Form.Get("buffer")
	bufId, err := strconv.ParseUint(bufHex, 16, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Get latest lines in reverse order.
	lines, err := weechatConn.BufferData(bufId, -256, "date,message")
	revlines := make([]weechat.LineData, 0, 256)
	for i := range lines {
		revlines = append(revlines, lines[len(lines)-1-i])
	}
	err = linesTpl.Execute(w, revlines)
	if err != nil {
		log.Printf("template error: %s", err)
	}
}

func handleWebsocket(conn *ws.Conn) {
}
