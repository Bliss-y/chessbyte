package main

import (
	"fmt"
	"net"
	"strconv"
)

const FEN = "rnbqkbnr";

type Actor struct {
    connection *PlayerConnection;
    next *Actor;
}

type PlayerConnection struct {
    connection net.Conn;
    channel chan string;
    id int;
    rating int;
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
    winner int;
}

var game_side = [2]string{"white","black"};

var side_char = [2]int{int('A'), int('a')};
var piece_char = [2]string{"rnbqk", "RNBQK"};

func getPiece(file int, rank int, fen *[68]byte) byte{
    return fen[rank * 8 + file];
}

func (g *Game) broadCast(message string) {
    for _, x := range g.players {
        if (x.connection != nil && !sendWebscoketMessage(x.connection , message)) {
            // TODO idk what to do here ngl
        }
    }
}

func (g *Game) initializeGame(conn net.Conn){
    g.players[0] = &PlayerConnection{conn, make(chan string),0 , 0}
    g.playerCount = 1; 
    for i,_ := range g.currentfen {
    g.currentfen[i] = ' ';
}
    for i, c:= range FEN {
        g.currentfen[i] = byte(c) - ('a' - 'A'); 
        g.currentfen[56+i] = byte(c)
    }
    g.turn = 0;
}

func (g *Game) reset() {
    for _,p := range g.players {
    if p.connection != nil {
        closeSocket(p.connection);
    }
    }
    g.players[0] = nil
    g.players[1] = nil
    g.playerCount = 0;
    g.status = -2;
}

func pieceSide(piece byte) int{
    if piece == ' ' {
        return -1;
    }
    if piece >= 'a' && piece <= 'z' {
        return 1;
    }
    return 0;
}

//message len = 7; message format [filechar a-h][rankchar 1-8][action][filechar a-h][rankchar 1-8] prev position to new pos
func (g *Game) update(message string, side int){
    nfen := g.currentfen;
    if g.turn != side {
        sendWebscoketMessage(g.players[side].connection, "decline, wrong turn");
        return;
    }
    prev_file := int(message[0]) - int('a'); 
    prev_rank, err := strconv.Atoi(message[1:2]);
    if err != nil || prev_rank >=8 || prev_file >=8{
        sendWebscoketMessage(g.players[side].connection, "decline, invalid prevrank or prev_file");
        return;
    }
    action := message[2];
    nfile := int(message[3]) - int('a');
    nrank, err := strconv.Atoi(message[4:5]);
    if err != nil || nrank > 7 || nfile > 7 {
        sendWebscoketMessage(g.players[side].connection, "decline, invalid nrank or nfile");
        return;
    }
    piece := getPiece(prev_file, prev_rank, &nfen);
    fmt.Println("piece ",piece, prev_file, prev_rank);
    if pieceSide(piece) != side {
        sendWebscoketMessage(g.players[side].connection, "decline, wrong piece");
        return;
    }
    opp_side := []int{1,0}
    if action == 'x' && pieceSide(getPiece(nfile, nrank, &nfen)) != opp_side[side]{
        sendWebscoketMessage(g.players[side].connection, "decline, wrong action piece");
        return;
    }
    // TODO: check the values;
    switch piece {
        case 'r','R':
            
        case 'n','N':
        case 'b', 'B':
        case 'q','Q':
        case 'k','K':
    }
    // TODO: check the integrity of the move i.e if the move makes king in check;
    nfen[nrank * 8 + nfile] = piece;
    nfen[prev_rank * 8 + prev_file] = ' ';
    g.broadCast(string(nfen[:]));
    g.currentfen = nfen;
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
        for i,p := range game.players {
            select {
                case message := <- p.channel:
                    if message == "Connection Closed!" ||  len(message) == 0 {
                        fmt.Println(message);
                        // someone disconnected.... TODO: wait for reconnect
                        if i == 1 {
                            game.winner = 0;
                            p.connection = nil;
                            game.broadCast("game end 1 0");
                        } else {
                            game.winner = 1;
                            p.connection = nil;
                            game.broadCast("game end 0 1");
                        }
                        game.status = -2;
                    } else {
                        game.update(message, i);
                    }
                default:
                    continue;
            } 
        }
    }
    game.reset();
}

