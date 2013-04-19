package main

import (
	"code.google.com/p/go.net/websocket"
	"encoding/json"
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
		err := websocket.JSON.Receive(ws, &msg)
		if err != nil {
			fmt.Println("rec err")
			break
		}
		fmt.Println(msg)

		termsStmt, serr := db.Prepare("SELECT full_address FROM dor_parcels where full_address like $1 order by full_address limit $2;")
		if serr != nil {
			fmt.Println(serr)
			return
		}

        rows, serr := termsStmt.Query(msg.Data+"%", 10)
        if serr != nil {
            fmt.Println(serr)
            return
        }

		results := []Address{}

		for rows.Next() {
			var addr string
			var result Address

			rows.Scan(&addr)
			result.Full = addr
			results = append(results, result)
		}

		msg.Event = "multiple"
		b, err := json.Marshal(results)
		if err != nil {
			fmt.Println(err)
			break
		}
		msg.Data = string(b)

		err = websocket.JSON.Send(ws, msg)
		if err != nil {
			fmt.Println("send err")
			break
		}
	}

	fmt.Println("closed socket")
}

func IndexHandler(c http.ResponseWriter, req *http.Request) {
	indexTmpl.Execute(c, req.Host)
}

func main() {
	port := "8989"
	if len(os.Args) == 2 {
		port = os.Args[0]
	}

	// Open the database connection pool for use by all socket connections
	db, dberr := sql.Open("postgres", "user=xxxxx dbname=xxxxx")
	if dberr != nil {
		fmt.Println(dberr)
		return
	}

	http.Handle("/echo/", DbWebsocketServer(JsonServer, db))
	http.Handle("/client/", http.StripPrefix("/client/", http.FileServer(http.Dir("client"))))
	http.HandleFunc("/", IndexHandler)

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
