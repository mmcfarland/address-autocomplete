package main

import (
    "fmt"
    "net/http"
    "os"
    "text/template"
    "code.google.com/p/go.net/websocket"
)

var indexTmpl = template.Must(template.ParseFiles("client.html"))

// Is there a way to force-expose lower case members (for javascript conventions)?
type JsonMsg struct {
    Event string
    Data string
}

func JsonServer(ws *websocket.Conn) {
    var msg JsonMsg
    fmt.Println("serving socket")
    for {
        err := websocket.JSON.Receive(ws, &msg)
        if err != nil {
            fmt.Println("rec err")
            break
        }
        fmt.Println(msg)
        msg.Event = "single"
        msg.Data = "I heard: "+msg.Data
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

    http.Handle("/echo/", websocket.Handler(JsonServer))
    http.Handle("/client/",  http.StripPrefix("/client/", http.FileServer(http.Dir("client"))))
    http.HandleFunc("/", IndexHandler)
    

    err := http.ListenAndServe(":"+port, nil)
    if err != nil {
        panic("ListenAndServe: " + err.Error())
    }
}
