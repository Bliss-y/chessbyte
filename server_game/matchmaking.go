package main

import (
    "strings"
)

// I may change this to a balanced tree instead
type MatchFindingPlayer struct {
    player *PlayerConnection;
    finding_tick int;
    idle int;
    next *MatchFindingPlayer;
    gameType string;
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
        if temp.player.ws.authToken == id {
            prev.next = temp.next;
            break;
        }
        prev = temp;
        temp = temp.next;
    }
    return temp;
}

func initMatchMakingPool() MatchMakingPool{
    pool := MatchMakingPool{nil, nil, make(map[string]*MatchFindingPlayer),0,0,0,make(chan *MatchFindingPlayer), make(chan string)};
    go pool.runMatchMaking();
    return pool;
}

func (p *MatchFindingPlayer) findPair() *MatchFindingPlayer {
    temp := p;
    for temp != nil {
        diff := p.player.rating - temp.player.rating;
        if diff * diff < 50 * 50 {
            return temp;
        }
        temp = temp.next;
    }
    return nil;
}

func (m *MatchMakingPool) runMatchMaking(){
    for {
        select {
            case message :=  <-m.messages: {
                if m.players == nil {
                    m.players = message;
                }
            }
            case message := <- m.signals:{
                commands := strings.Split(message, " ");
                if commands[0] == "removePlayer" {
                    m.players.remove(commands[1]);
                    m.finders.remove(commands[1]);
                    delete(m.playing, commands[1]);
                }
            }
            default: {
                temp := m.players;
                for temp != nil {
                    select {
                        case pl_message:= <-temp.player.ws.channel: {
                            if pl_message == "" || temp.player.ws.connection == nil {
                                delete(m.playing, temp.player.ws.authToken);
                                m.players.remove(temp.player.ws.authToken);
                            }
                            if pl_message == "find" {
                                m.players.remove(temp.player.ws.authToken)
                                if m.finders == nil {
                                    m.finders = temp;
                                } else {m.finders.add(temp)}
                            }
                        }
                        default: 
                    }
                    temp = temp.next;
                }
                temp = m.finders;
                for temp != nil {
                        if temp.player.ws.closed {
                            m.players.remove(temp.player.ws.authToken);
                            temp = temp.next;
                            continue;
                        }
                        select {
                        case pl_message:= <- temp.player.ws.channel: {
                            if pl_message == "" || temp.player.ws.connection == nil {
                                delete(m.playing, temp.player.ws.authToken);
                                m.finders.remove(temp.player.ws.authToken);
                            }
                            if pl_message == "stop" {
                                m.finders.remove(temp.player.ws.authToken)
                                if m.finders == nil {
                                    m.players = temp;
                                } else {m.players.add(temp)}
                            }
                        }
                        default: {
                            pair := m.players.findPair();
                            if pair != nil {
                                // TODO: Create a Game here;
                                game := &Game {
                                }
                                go gameLoop(game);
                            }
                        }
                    }
                    temp = temp.next;
                }
            }
        }
    }
}

func (m *MatchMakingPool) add(player *PlayerConnection) {
    
}
