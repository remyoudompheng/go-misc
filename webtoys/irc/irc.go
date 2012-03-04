package irc

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/textproto"
	"os"
	"strings"

	"code.google.com/p/go.net/websocket"
	"encoding/json"
)

var logger = log.New(os.Stderr, "irc ", log.LstdFlags|log.Lshortfile)

func init() {
	http.HandleFunc("/irc", home)
	http.Handle("/irc/ws", websocket.Handler(connect))
	logger.Printf("registered irc at /irc and /irc/ws")
}

const (
	nick    = "remy_web"
	server  = "chat.freenode.net:6667"
	channel = "#arch-fr-off"
)

type Client struct {
	*textproto.Conn
	lines chan string
}

func NewClient(server, nick string) (c *Client, err error) {
	conn, err := textproto.Dial("tcp", server)
	if err != nil {
		return
	}
	// RFC 2812, 3.1.2
	conn.Cmd("NICK %s", nick)
	// RFC 2812, 3.1.3
	conn.Cmd("USER %s %d %s :%s", "webtoy", 0, "*", "Anonymous Guest")
	c = &Client{
		Conn:  conn,
		lines: make(chan string, 8),
	}
	go c.ReadLines()
	return c, nil
}

func (cli *Client) Send(cmd string, args ...interface{}) {
	cli.Cmd("%s %s", cmd, fmt.Sprint(args))
}

func (cli *Client) ReadLines() error {
	defer close(cli.lines)
	for {
		line, err := cli.ReadLine()
		if err != nil {
			return err
		}
		cli.lines <- line
	}
	panic("unreachable")
}

func parseIrcLine(line string) (prefix, command string, args []string) {
	if line == "" {
		logger.Printf("cannot parse %q", line)
		return
	}
	var hasPrefix bool
	if line[0] == ':' {
		hasPrefix = true
	}
	var items []string
	colon := strings.Index(line[1:], ":")
	if colon >= 0 {
		// long arg
		items = strings.Fields(line[:colon+1])
		items = append(items, line[colon+2:])
	} else {
		items = strings.Fields(line)
	}
	if len(items) < 2 {
		logger.Printf("cannot parse %q", line)
		return
	}
	if hasPrefix {
		return items[0][1:], items[1], items[2:]
	}
	return "", items[0], items[1:]
}

func formatIrcLine(prefix, command string, args []string) string {
	switch command {
	case "PRIVMSG":
		if i := strings.Index(prefix, "!"); i >= 0 {
			prefix = prefix[:i]
		}
		if len(args) != 2 {
			return fmt.Sprintf("%.10s: %v", prefix, args)
		}
		return fmt.Sprintf("(%.10s) %.10s: %v", args[0], prefix, args[1])
	}
	return fmt.Sprintf("/%s %v", command, args)
}

type Form struct {
	Nick, Chan, Serv string
}

func loggedmessage(conn *websocket.Conn, format string, args ...interface{}) {
	logger.Printf("/irc/ws: "+format, args...)
	websocket.Message.Send(conn, fmt.Sprintf(format, args...))
}

func connect(conn *websocket.Conn) {
	logger.Printf("websocket from %s", conn.RemoteAddr())
	defer conn.Close()
	var form []byte
	var f Form
	if err := websocket.Message.Receive(conn, &form); err != nil {
		return
	}
	if err := json.Unmarshal(form, &f); err != nil {
		loggedmessage(conn, "invalid request: %s (%s)", form, err)
		return
	}
	loggedmessage(conn, "opening connection to %s for %s", f.Serv, f.Nick)
	client, err := NewClient(f.Serv, f.Nick)
	if err != nil {
		websocket.Message.Send(conn, "connection error: "+err.Error())
		return
	}
	defer logger.Printf("closing connection to %s for %s", f.Serv, f.Nick)
	defer client.Close()
	defer client.Cmd("QUIT :%s", "client left.")
	logger.Printf("joining channel %s", f.Chan)
	client.Cmd("JOIN %s", f.Chan)
	for line := range client.lines {
		websocket.Message.Send(conn, formatIrcLine(parseIrcLine(line)))
	}
}

const ircTemplate = `<!DOCTYPE html>
<html>
<head>
	<title>IRC Chat</title>
	<meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
	<script type="text/javascript" src="http://ajax.googleapis.com/ajax/libs/jquery/1.4.2/jquery.min.js"></script>
	<script type="text/javascript">
  	$(function() {
		var conn;
		var WS = window["WebSocket"] ? WebSocket : MozWebSocket;

		function connect() {
			$("#lines").empty();
			var params = {};
			params.Nick = $("#nick").val();
			params.Chan = $("#channel").val();
			params.Serv = $("#server").val();
			conn = new WS("ws://{{ $ }}/irc/ws");
			conn.onopen = function () {
				conn.send(JSON.stringify(params));
			};
			conn.onmessage = display;
			return false;
		};

		function display(msg) {
			$("#lines").append($("<li/>").text(msg.data));
		};

		$("#form").submit(connect);

		$.unload(function() {
			if (conn) {
				conn.close();
			}
		});
	});
	</script>
	<style type="text/css">
		ul#lines {
			list-style: none;
		}
		ul#lines li {
			font-family: monospace, Courier;
		}
	</style>
</head>
<body>
	<h1>IRC Chat</h1>

	<form id="form">
		Nick:
		<input type="text" id="nick" size="32" value="guest">
		Channel:
		<input type="text" id="channel" size="32" value="#arch-fr-off"><br/>
		Server:
		<input type="text" id="server" size="32" value="chat.freenode.net:6667">
		<input type="submit" value="Connect">
	</form>

	<ul id="lines">
	</ul>
</body>
</html>
`

var ircTpl = template.Must(template.New("irc").Parse(ircTemplate))

func home(resp http.ResponseWriter, req *http.Request) {
	logger.Printf("GET %s from %s", req.URL, req.RemoteAddr)
	err := ircTpl.Execute(resp, req.Host)
	if err != nil {
		logger.Printf("error: %s", err)
	}
}
