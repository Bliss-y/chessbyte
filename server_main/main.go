package main

import (
    "fmt"
    "net"
    "crypto/sha1"
    "bufio"
    "net/http"
    "io"
    "bytes"
    "encoding/base64"
)

const FEN = "rnbqkbnr";

type PlayerConnection struct {
    connection net.Conn;
    channel chan string;
    id int;
    rating int;
}


type Actor struct {
    connection *PlayerConnection;
    next *Actor;
}

// Game is just a room.
type Game struct {
    players [2]*PlayerConnection;
    viewers *Actor;
    playerCount int;
    viewerCount int;
    currentfen [68]byte;
    lastmove int;
    turn int;
    gameType int;
    status int; //-2 idle, -1 -> finding, 0 white wins, 1 black wins
}

func (g *Game) initializeGame(conn net.Conn){
    g.players[0] = &PlayerConnection{conn, make(chan string),0 , 0}
    g.playerCount = 1; 
    for i, c:= range FEN {
        g.currentfen[i] = byte(c) - 47; 
        g.currentfen[56+i] = byte(c)
    }
    g.turn = 0;
}

func (g *Game) reset() {
    g.players[0] = nil
    g.players[1] = nil
    g.playerCount = 0;
    g.status = -2;
}

func gameLoop(game *Game) {
    for game.status == -1 {
    }
    if (game.status == 2) {
        fmt.Println("Partner found!! ")
        for _,e := range game.players {
            go readWebsocket(e.connection,e.channel);
        }
    }
    for game.status == 2 {
        for _,p := range game.players {
            select {
                case message := <- p.channel:
                    fmt.Println(message);
                default:
                    continue;
            } 
        }
    }
}


var games = [1]*Game{};

func sendWebscoketMessage() {
     
}

func readWebsocket(conn net.Conn, message chan string) {
    buffer := make([]byte, 50); 
    for {
        n, err := conn.Read(buffer);
        if err == io.EOF {
            close(message)
            break;
        } else if err != nil {
          close(message);
          break;
        }

        // do the message parsing
        fin := 0;
        opcode :=0;
        fin = int((buffer[0] >> 7) & 1)
        opcode = int(buffer[0] & 0x0F);
        mask := (buffer[1] >> 7) & 1;
        payload_length := buffer[1] & 0x7F; // last 7 bits
        current := 2;
        if (payload_length == 126) {
            // read next 16 bytes for payload lengthh
            payload_length = buffer[2];
            current = 4;
        } else if (payload_length == 127) {
            // read next 64 bytes for payload length
            payload_length = recb[2];
            current = 10;
        }
        masking_key := []byte{buffer[current + 0], buffer[current + 1], buffer[current + 2], buffer[current + 3]};
        current += 4;
        for i := 0; i < payload_length; i++ {
            reciever[i] = rune(buffer[i + current] ^ masking_key[i % 4]);
            printf("\n");
            printf("%c", reciever[i]);
            printf("\n");
        }

    }
}

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
            // do the websocket handshake
            if req.Method != "GET" {
                conn.Close();   
                return;
            }
            connection_header, exists:= req.Header["Upgrade"];
            if !exists || connection_header[0] != "websocket" {
                conn.Close();   
                return;
            }
            websocketkey, exists := req.Header["Sec-Websocket-Key"];
            websocketkey[0] += "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

            hash := sha1.New()
	        // Write the string data to the hash object
            hash.Write([]byte(websocketkey[0]));
	        // Get the final hash as a byte slice
	        hashBytes:= hash.Sum(nil)
            encoded:= base64.StdEncoding.EncodeToString(hashBytes);    
            var responsebytes bytes.Buffer;
            fmt.Fprintf(&responsebytes, "HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: %s\r\n\r\n", encoded);
            fmt.Println(responsebytes);
            n, err := conn.Write([]byte(responsebytes.String()));
            fmt.Println(n)
            if err != nil {
                conn.Close();
                return
            }
            // check for existing games first
            for i,game:= range games {
                if game == nil {
                   continue;
                }
                if game.status == -1 {
                     // join the game
                     fmt.Println("Joining the existing game")
                     game.status = 2
                     game.players[1] = &PlayerConnection{conn, make(chan string), 0, 0};
                     game.playerCount = 2;
                     games[i] = game
                     return;
                }
            }
            // check for idle games
            for i,game:= range games {
                if game == nil {game = &Game{};
                game.status = -2}
                if game.status == -2 {
                     game.initializeGame(conn)
                     fmt.Println("Creating the new game")
                     // join the game
                     game.status = -1
                     games[i] = game
                     go gameLoop(games[i]);
                     return;
                }
            }
            fmt.Println("Game not found ")
            conn.Close();
            return
        }
        conn.Close()
}


