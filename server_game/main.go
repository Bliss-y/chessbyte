package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
)

const SERVER_TOKEN = "SERVER_TOKEN_FORNOW"
const SERVER_port = 4000;
var games = [30]*Game{};

func main(){
    ln, err := net.Listen("tcp", ":3000") 
    if (err != nil){
        fmt.Println("could not start the server. check the port")
        fmt.Println(err)
        return;
    }
    matchMakingpoool := initMatchMakingPool();
    onConnection := func(conn *WebSocketConnection){
        fmt.Println("new connection!")
        player := &PlayerConnection{};
        player.ws = conn;
        res, err := http.Get(
            fmt.Sprintf("http://localhost:4000/AuthPlayer?token=%v",
            conn.authToken));

        if err != nil {
            fmt.Println("user not logged in, didn't work");
            conn.connection.Close();
        }
        resbytes := make([]byte, 0);
        res.Body.Read(resbytes)
        type resFromReq struct {
            name string;
            id string;
            auth bool;
            rating uint;
        }
        p := resFromReq{}
        if json.Unmarshal(resbytes, &p) != nil {
            panic("could not parse data from there");
        }
        player.rating = p.rating;
        player.id = p.id;
        matchMakingpoool.add(player);
    }
    for {
        conn, err := ln.Accept()
        if err != nil {
            continue
        }
        go handleRequest(conn, onConnection);
    }
}
