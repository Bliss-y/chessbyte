package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"time"
)
const FPS = 100; // 10 fps.
const FEN = "rnbqkbnr";
const GAME_STATUS_RUNNING = 3;
const GAME_STATUS_WHITE_WIN = 0;
const GAME_STATUS_BLACK_WIN = 1;
const GAME_STATUS_DRAW = 2;

type Actor struct {
    connection *PlayerConnection;
    channel chan string;
    next *Actor;
}

// Game is just a room.
type Game struct {
    players [2]*PlayerConnection;
    viewers *Actor;
    playerCount int;
    timeout int;
    viewerCount int;
    reason string;
    currentfen [68]byte;
    timed bool;
    increment int;
    tick int;
    game_duration int; // in ticks
    lastmove int;
    turn int;
    gameType int;
    status int; //-2 idle, -1 -> finding, 0 white wins, 1 black wins, 2 draw, 3 running
    winner int;
}

var game_side = [2]string{"white","black"};
var side_char = [2]int{int('A'), int('a')};
var piece_char = [2]string{"rnbqk", "RNBQK"};
var opp_side = [2]int{1,0};

func getPiece(file int, rank int, fen *[68]byte) byte{
    return fen[rank * 8 + file];
}

func (g *Game) broadCast(message string) {
    for _, x := range g.players {
        if (!x.ws.closed && !x.ws.sendWebscoketMessage(message)) {
            // TODO idk what to do here ngl
        }
    }
}

func (g *Game) initializeGame(player1 *PlayerConnection, player2 *PlayerConnection, gameType string){
    g.turn = 0;
    g.timeout = 300; // 30 seconds of idle = game over
    g.playerCount = 2; 
    p1_side := rand.Intn(2);
    if p1_side == 0 {
        player1.side = 0;
        player2.side = 1;
        g.players = [2]*PlayerConnection{player1, player2};
    }else {
        player2.side = 0;
        player1.side = 0;
        g.players = [2]*PlayerConnection{player2, player1};
    }
    switch gameType {
        case "blit": {
            g.timed = true;
            g.game_duration = 10*60*5; //5 minutes blitz
        }
        case "10min": {
                g.timed = true;
                g.game_duration = 10*60*10; //5 minutes blitz
        }
        case "notTimed": {
            g.timed = false;
        }
    }
    for _,p := range g.players {
        p.timer = g.game_duration;
        p.timeout = g.timeout;
    }
}

func (g *Game) reset() {
    for _,p := range g.players {
        p.destroy();
    }
    g.players[0] = nil
    g.tick = 0;
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
func (g *Game) update(dt int64){
    if g.status != GAME_STATUS_RUNNING {
        return;
    }
    for _,p := range g.players {
        p.update(dt, g);
    }
        var responsebytes bytes.Buffer;
        fmt.Fprintf(&responsebytes, "%s %v %v %v %v", string(game.currentfen[:]), game.game_duration, game.turn, game.tick)
        g.broadCast(responsebytes.String());
        g.tick++;
}

func gameLoop(game *Game) {
    // TODO: let game loop only handle while playing mode
    for game.status == -1 {
    }
    if (game.status == GAME_STATUS_RUNNING) {
        fmt.Println("Starting game")
        for _,e := range game.players {
            go readWebsocket(e.connection,e.channel);
        }
    }
    start := time.Now();
    for game.status == GAME_STATUS_RUNNING {
        end := time.Now();
        dt := end.Sub(start).Milliseconds() 
        // Every tick is 100ms
        if dt < FPS {
            continue;
        }
        game.update(dt);
        start = time.Now();
    }
    game.reset();
}
