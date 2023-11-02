package main

import (
    "fmt"
    "net"
    "bufio"
    "net/http"
    "strings"
)

func main(){
    ln, err := net.Listen("tcp", ":3000") 
    if (err != nil){
        fmt.Println("could not start the server. check the port")
        fmt.Println(err)
        return;
    }

    for {
        conn, err := ln.Accept()
        if err != nil {
            continue
        }
        go handleRequest(conn);
    }
}

func handleRequest(conn net.Conn){
        req,err := http.ReadRequest(bufio.NewReader(conn))
        if err != nil{
            conn.Close();
            return
        }
        if req.URL.Path == "/game" {
            // This is where You do the web socket stuffs 
            fmt.Println(req);
            fmt.Fprintf(conn,"HTTP/1.1 200\r\n\r\nHello World!!");
            conn.Close()
            return
        }
        if strings.Contains(req.URL.Path, "/"){
        
        }
        fmt.Fprintf(conn,"HTTP/1.1 404\r\n\r\n");
        conn.Close()
}


