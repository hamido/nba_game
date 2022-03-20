package main

import (
	"bytes"
	"flag"
	"github.com/gorilla/websocket"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

const AttackTime = time.Second * 24
const GameTime = time.Second * 240

var addr = flag.String("addr", "localhost:8080", "http service address")

var upgrader = websocket.Upgrader{} // use default options

func stats(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	updates := make(chan []Stats)
	defer c.Close()
	var numberOfMatches int

	for {
		_, message, _ := c.ReadMessage()
		numberOfMatches, _ = strconv.Atoi(string(message))
		break
	}

	go Week(numberOfMatches, updates)

	for {
		select {
		case u := <-updates:
			buffer := new(bytes.Buffer)
			err := statsTemplate.Execute(buffer, u)
			if err != nil {
				panic(err)
			}
			err = c.WriteMessage(1, buffer.Bytes())
			if err != nil {
				//log.Println("write:", err)
			}
		}
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	homeTemplate.Execute(w, nil)
}

func main() {
	http.HandleFunc("/stats", stats)
	http.HandleFunc("/", home)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

var statsTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
</head>
<title>Score Board</title>
<style>
	table, th, td {
		border:1px solid black;
	}
</style>
<body>
	<script>
		window.addEventListener("load", function(evt) {
			var ws = new WebSocket("ws://localhost:8080/stats");
			ws.onmessage = function(evt) {
				var node = document.getElementById('content');
				node.innerHTML = evt.data;
			}
		})
	</script>
	<div id="content">
		{{range .}}
		<h2>{{.Name}} | Attack Count: {{.AttackCount}} | Total Score: {{.TotalScore}}</h2>
		<table style="width:100%">
			<tr>
				<th>Name</th>
				<th>Assists</th>
				<th>Points 2 Rate</th>
				<th>Point 3 Rate</th>
			</tr>
			{{ range $key, $value := .PlayerStats }}
			<tr>
				<th>{{$key}}</th>
				<th>{{$value.Assists}}</th>
				<th>{{$value.Points2Rate}}</th>
				<th>{{$value.Points3Rate}}</th>
			</tr>
			{{end}}
		</table>
		{{end}}
	</div>
</body>
</html>
`))

var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8">
</head>
<title>Score Board</title>
<body>
	<script>
		window.addEventListener("load", function(evt) {
			var ws = new WebSocket("ws://localhost:8080/stats");
			ws.onmessage = function(evt) {
				var node = document.getElementById('content');
				node.innerHTML = evt.data;
			}

			document.getElementById("start").onclick = function(evt) {
				if (!ws) {
					return false;
				}
        		ws.send(input.value);
        		return false;
    		};
		})
	</script>
	<div id="content">
		<form>
			<p><input id="input" type="number" value="2">
			<button id="start">Start</button>
		</form>
	</div>
</body>
</html>
`))

func Random() bool {
	rand.Seed(time.Now().UnixNano())
	return rand.Float32() < 0.5
}
