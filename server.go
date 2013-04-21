package main

import (
	"code.google.com/p/go.net/websocket"
	"database/sql"
	"encoding/json"
	"flag"
	_ "github.com/bmizerany/pq"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/template"
)

const l = "server.log"

var f, err = os.OpenFile(l, os.O_APPEND|os.O_CREATE, 0666)
var lf = log.New(f, "", log.Ldate|log.Ltime)

var port = flag.Int("port", 8989, "Port")

var indexTmpl = template.Must(template.ParseFiles("client.html"))

type JsonMsg struct {
	Event string
	Data  string
}
type Address struct {
	Full string
}

func DbWebsocketServer(fn func(ws *websocket.Conn, db *sql.DB), db *sql.DB) websocket.Handler {
	return func(ws *websocket.Conn) {
		fn(ws, db)
	}
}

func JsonServer(ws *websocket.Conn, db *sql.DB) {
	var msg JsonMsg
	for {
		if err := websocket.JSON.Receive(ws, &msg); err != nil {
			lf.Println(err)
			break
		}

		sql := "SELECT full_address FROM dor_parcels where ts_f_address @@ to_tsquery($1) order by full_address limit $2;"
		termsStmt, err := db.Prepare(sql)
		if err != nil {
			lf.Println(err)
			break
		}

		// AND together all search tokens with the last using a prefix wildcard
		tsquery := strings.Replace(strings.Trim(msg.Data, " "), " ", " & ", -1) + ":*"
		rows, err := termsStmt.Query(tsquery, 10)
		if err != nil {
			lf.Println(err)
			break
		}

		var results []Address

		for rows.Next() {
			var addr string
			var result Address

			rows.Scan(&addr)
			result.Full = addr
			results = append(results, result)
		}

		if b, err := json.Marshal(results); err != nil {
			lf.Println(err)
			break
		} else {
			msg.Event = "multiple"
			msg.Data = string(b)
		}

		if err := websocket.JSON.Send(ws, msg); err != nil {
			lf.Println(err)
			break
		}
	}
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	indexTmpl.Execute(w, r.Host)
}

func main() {

	flag.Parse()

	// Open the database connection pool for use by all socket connections
	db, err := sql.Open("postgres", "user=xxxx dbname=xxxxx")
	if err != nil {
		lf.Println(err)
		return
	}
	defer db.Close()

	http.Handle("/suggest/", DbWebsocketServer(JsonServer, db))
	http.Handle("/client/", http.StripPrefix("/client/", http.FileServer(http.Dir("client"))))
	http.HandleFunc("/", IndexHandler)

	if err := http.ListenAndServe(":"+strconv.Itoa(*port), nil); err != nil {
		lf.Panic("ListenAndServe: " + err.Error())
	} else {
		lf.Println("Listening on port: " + strconv.Itoa(*port))
	}
}
