package main

import (
	"math"
	"strings"
	"time"
)

// TODO: CLEAR ALL DEFAULTS IN SELECT STATEMENT

// NOTE: I may change this to a balanced tree instead
type MatchFindingPlayer struct {
    player *PlayerConnection;
    finding_tick uint;
    idle int;
    next *MatchFindingPlayer;
    gameType string;
    status string; // playing, finding, idle, disconnected
    game *Game;
}

// MatchMakingPool should contain all the players at all times
type MatchMakingPool struct {
    players *MatchFindingPlayer;
    finders *MatchFindingPlayer;
    playing map[string]*MatchFindingPlayer; //i'm thinking of putting in authkey as mapkey
    playersn int;
    findersn int;
    playingn int;
    messages chan *MatchFindingPlayer;
    signals chan string;
}

func (p *MatchFindingPlayer) add(np *MatchFindingPlayer){
    if p == nil {
        return;
    }
    x := p;
    for x.next != nil {
        x = x.next;
    }
    x.next = np;
}

func (p *MatchFindingPlayer) remove(id string ) *MatchFindingPlayer{
    if p == nil {
        return nil;
    }
    temp := p;
    prev := temp;
    for temp.next != nil {
        if temp.player.id == id {
            prev.next = temp.next;
            break;
        }
        prev = temp;
        temp = temp.next;
    }
    return temp;
}

func initMatchMakingPool() MatchMakingPool{
    pool := MatchMakingPool{nil, nil, 
        make(map[string]*MatchFindingPlayer),
        0,0,0,
        make(chan *MatchFindingPlayer), make(chan string)};
    go pool.runMatchMaking();
    return pool;
}

func (p *MatchFindingPlayer) findPair() *MatchFindingPlayer {
    temp := p;
    for temp != nil {
        diff := p.player.rating - temp.player.rating;
        // What the fuck????? GO..?
        factor := uint(math.Floor(math.Exp2(float64(p.finding_tick/5))));
        if diff * diff <= ((50 * 50) + factor){
            return temp;
        }
        temp = temp.next;
    }
    p.finding_tick++;
    return nil;
}

func (m *MatchMakingPool) RemovePlayer(identifier string){
    m.players.remove(identifier);
    m.finders.remove(identifier);
    m.playing[identifier].player.ws.Close();
}

func (m *MatchMakingPool) runMatchMaking(){
    startTime := time.Now()
    overflow := int64(0);
    for {
        select {
            case message :=  <-m.messages: {
                if m.players == nil {
                    m.players = message;
                }
            }
            case message := <- m.signals:{
                commands := strings.Split(message, " ");
                if len(commands) == 2 && commands[0] == "removePlayer" {
                    m.RemovePlayer(commands[1]);
                }
            }
            }
        diff := time.Now().Sub(startTime).Milliseconds();
        if diff < 100-overflow {
            continue;
        }
        // For all idle players
        temp := m.players;
        for temp != nil {
            select {
                case pl_message:= <-temp.player.ws.channel: {
                    if pl_message ==  "" {
                        delete(m.playing, temp.player.id);
                        m.players.remove(temp.player.id);
                        break;
                    }
                    cmds := strings.Split(pl_message, " ")
                    if len(cmds) < 2 {break;}
                    switch cmds[0] {
                    //TODO: Somwhow have gametype here
                    case "find": {
                        m.players.remove(temp.player.id)
                        //TODO: check for the correct type
                        temp.gameType = cmds[1];
                        if m.finders == nil {
                            m.finders = temp;
                        } else {
                            m.finders.add(temp)
                        }
                    }
                }
            }
            }
            temp = temp.next;
        }
        // for all finding players
        temp = m.finders;
        for temp != nil {
                if temp.player.ws.closed {
                    m.players.remove(temp.player.id);
                    temp = temp.next;
                    continue;
                }
                select {
                case pl_message:= <- temp.player.ws.channel: {
                    switch pl_message {
                        case "": {
                            delete(m.playing, temp.player.id);
                            m.finders.remove(temp.player.id);
                        }
                        case "stop": {
                            m.finders.remove(temp.player.id);
                            if m.finders == nil {
                                m.players = temp;
                            } else {m.players.add(temp)}
                        }
                }
            }
            default: {
                pair := m.players.findPair();
                if pair != nil {
                    // TODO: Create a Game here;
                    game := &Game {
                    }
                    game.initializeGame(temp.player, pair.player, "blit");
                    go gameLoop(game);
                }
                }
            }
            temp = temp.next;
        }
        overflow = time.Now().Sub(startTime).Milliseconds() - 100;
        startTime =  time.Now()
        }
}

func (m *MatchMakingPool) add(player *PlayerConnection) {
    x,exists :=m.playing[player.ws.authToken]
    if exists && x.status == "playing" {
        player.ws.Close();
        return;
    }
    mfPlayer := &MatchFindingPlayer{player, 0, 0, nil, "", "idle", nil}
    m.messages <- mfPlayer;
}
