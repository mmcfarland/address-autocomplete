package main

import (
    "fmt"
    "net/http"
    "os"
    "text/template"
    "code.google.com/p/go.net/websocket"
)

var indexTmpl = template.Must(template.ParseFiles("client.html"))

func EchoServer(ws *websocket.Conn) {
    var msg string
    fmt.Println("serving")
    err := websocket.Message.Receive(ws, &msg)
    if err != nil {
        fmt.Println("rec err")
    }
    fmt.Println(msg)
    err = websocket.Message.Send(ws, "I heard: " + msg)
    if err != nil {
        fmt.Println("send err")
    }
}

func IndexHandler(c http.ResponseWriter, req *http.Request) {
    indexTmpl.Execute(c, req.Host)
}

func main() {
    port := "8989"
    if len(os.Args) == 2 {
        port = os.Args[0]
    }

    http.Handle("/echo/", websocket.Handler(EchoServer))
    http.HandleFunc("/", IndexHandler)

    err := http.ListenAndServe(":"+port, nil)
    if err != nil {
        panic("ListenAndServe: " + err.Error())
    }
}
