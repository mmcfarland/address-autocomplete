package main

import (
	"code.google.com/p/go.net/websocket"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
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
	Event string `json:"event"`
	Data  string `json:"data"`
}
type ParcelMatch struct {
	Address  string `json:"address"`
	Owner1   string `json:"owner1"`
	Owner2   string `json:"owner2"`
	Pos      string `json:"pos"`
	ParcelId int    `json:"parcelId"`
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
			log.Println(err)
			break
		}

		sql := `SELECT parcelid, address, owner1, owner2, concat('[',ST_Y(pos),',', ST_X(pos),']') as coord
                FROM pwd_parcels 
                WHERE ts_address @@ to_tsquery($1) 
                ORDER BY full_address limit $2;`
		termsStmt, err := db.Prepare(sql)
		if err != nil {
			log.Println(err)
			break
		}
		// Any wildcards will get converted into prefeix wildcard
		term := strings.Replace(msg.Data, "*", ":*", -1)

		// AND together all search tokens with the last using a prefix wildcard
		tsquery := strings.Replace(strings.Trim(term, " "), " ", " & ", -1) + ":*"
		rows, err := termsStmt.Query(tsquery, 10)
		if err != nil {
			log.Println(err)
			fmt.Println(err)
			break
		}

		var results []ParcelMatch

		for rows.Next() {
			var result ParcelMatch

			rows.Scan(&result.ParcelId, &result.Address, &result.Owner1,
				&result.Owner2, &result.Pos)
			results = append(results, result)
		}

		if b, err := json.Marshal(results); err != nil {
			log.Println(err)
			break
		} else {
			msg.Event = "multiple"
			msg.Data = string(b)
		}

		if err := websocket.JSON.Send(ws, msg); err != nil {
			log.Println(err)
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
	db, err := sql.Open("postgres", "user=postgres dbname=dataviewer")
	if err != nil {
		log.Println(err)
		return
	}
	defer db.Close()

	http.Handle("/suggest/", DbWebsocketServer(JsonServer, db))
	http.Handle("/client/", http.StripPrefix("/client/", http.FileServer(http.Dir("client"))))
	http.HandleFunc("/", IndexHandler)

	if err := http.ListenAndServe(":"+strconv.Itoa(*port), nil); err != nil {
		log.Panic("ListenAndServe: " + err.Error())
	} else {
		log.Println("Listening on port: " + strconv.Itoa(*port))
	}
}
