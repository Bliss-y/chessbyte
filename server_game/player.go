package main

import (
    "fmt"
    "strconv"
)

type PlayerConnection struct {
    ws *WebSocketConnection;
    id int;
    side int;
    rating uint;
    timeout int; // timeout to hold the last time before last move.
    timer int;
}

func (p *PlayerConnection)init(side int, timer int, timeout int) {
    p.side = side;
    p.timeout = timeout;
    p.timer = timer;
}

func makePlayer(ws *WebSocketConnection, id int, rating uint) *PlayerConnection {
    player := &PlayerConnection{ws, id, -1, rating, 0,0};
    return player;
}

func (p *PlayerConnection) destroy() {
    p.ws.Close();
    p.side = 0;
}

// TODO: Kind of a Circular dependency here 
func (p *PlayerConnection) update(dt int64, g *Game) {
    channel := p.ws.channel;
    side := p.side;
    if g.timed && g.players[side].timer <= 0 {
        g.status = opp_side[side];
        g.reason = "timed out"
        return;
    }
    select {
    case message := <- channel: {
        if message == "resign" || g.players[side].timeout <=0{
            g.status = opp_side[p.side];
            g.reason = "too idle abandoned"
            return;
        }
        if g.turn != side {
            return;
        }
        g.players[side].timeout = g.timeout;
        if g.turn != side {
            return;
        }
        nfen := g.currentfen;
        prev_file := int(message[0]) - int('a'); 
        prev_rank, err := strconv.Atoi(message[1:2]);
        if err != nil || prev_rank >=8 || prev_file >=8{
            return;
        }
        action := message[2];
        nfile := int(message[3]) - int('a');
        nrank, err := strconv.Atoi(message[4:5]);
        if err != nil || nrank > 7 || nfile > 7 {
            return;
        }
        piece := getPiece(prev_file, prev_rank, &nfen);
        fmt.Println("piece ",piece, prev_file, prev_rank);
        if pieceSide(piece) != side {
            return;
        }
        if action == 'x' && pieceSide(getPiece(nfile, nrank, &nfen)) != opp_side[side]{
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
        // TODO: check the integrity of the move i.e if the move makes king in check and for checkmate on enemy king;
        nfen[nrank * 8 + nfile] = piece;
        nfen[prev_rank * 8 + prev_file] = ' ';
        g.broadCast(string(nfen[:]));
        g.currentfen = nfen;
    }
    default: { // idle
        }
    }
    if g.turn != side {
        return;
    }
    g.players[side].timeout--;
    p.timer--;
    return;
}

