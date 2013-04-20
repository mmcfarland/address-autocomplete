package main

import (
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"strings"
	"fmt"
	"net/http"
	"os"
	"text/template"

	"database/sql"
	_ "github.com/bmizerany/pq"
)

var indexTmpl = template.Must(template.ParseFiles("client.html"))

// Is there a way to force-expose lower case members (for javascript conventions)?
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
	fmt.Println("serving socket")
	for {
		if err := websocket.JSON.Receive(ws, &msg); err != nil {
			fmt.Println("rec err")
			break
		}

		sql := "SELECT full_address FROM dor_parcels where ts_f_address @@ to_tsquery($1) order by full_address limit $2;"
		termsStmt, err := db.Prepare(sql)
		if err != nil {
			fmt.Println(err)
			break
		}

		tsquery := strings.Replace(strings.Trim(msg.Data, " "), " ", " & ", -1) + ":*"
		fmt.Println(tsquery)
		rows, err := termsStmt.Query(tsquery, 10)
		if err != nil {
			fmt.Println(err)
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
			fmt.Println(err)
			break
		} else {
			msg.Event = "multiple"
			msg.Data = string(b)
		}	

		if err := websocket.JSON.Send(ws, msg); err != nil {
			fmt.Println("send err")
			break
		}
	}

	fmt.Println("closed socket")
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	indexTmpl.Execute(w, r.Host)
}

func main() {
	port := "8989"
	if len(os.Args) == 2 {
		port = os.Args[0]
	}

	// Open the database connection pool for use by all socket connections
	db, err := sql.Open("postgres", "user=xxxx dbname=xxxxx")
	if err != nil {
		fmt.Println(err)
		return
	}

	http.Handle("/suggest/", DbWebsocketServer(JsonServer, db))
	http.Handle("/client/", http.StripPrefix("/client/", http.FileServer(http.Dir("client"))))
	http.HandleFunc("/", IndexHandler)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
