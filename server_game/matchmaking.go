package main

import (
	"math"
    "fmt"
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

func (p *MatchFindingPlayer) add(pool **MatchFindingPlayer,np *MatchFindingPlayer){
    if *pool == nil {
        *pool = np;
        return;
    }
    np.next = (*pool)
    *pool = np;
    /*
    if p == nil {
        return;
    }
    x := p;
    i:=0;
    for x.next != nil {
        x = x.next;
        i++;
        fmt.Println(i)
    }
    x.next = np;
    */
}

func (p *MatchFindingPlayer) remove(pool **MatchFindingPlayer, id string ) *MatchFindingPlayer{
    temp := (*pool);
    var prev *MatchFindingPlayer = nil;
    for temp != nil  {
        if temp.player.id == id {
            if prev == nil{
                *pool = temp.next;
                return temp;
            }
            prev.next = temp.next
            return temp;
        }
        prev = temp;
        temp = temp.next;
    }
    return nil;
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
        if diff * diff - temp.finding_tick<= ((50 * 50) + factor){
            return temp;
        }
        temp = temp.next;
    }
    p.finding_tick++;
    return nil;
}

func (m *MatchMakingPool) RemovePlayer(identifier string){
    m.players.remove(&m.players,identifier);
    m.finders.remove(&m.players,identifier);
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
                }else {
                    m.players.add(&m.players,message);
                }
            }
            case message := <- m.signals:{
                commands := strings.Split(message, " ");
                if len(commands) == 2 && commands[0] == "removePlayer" {
                    m.RemovePlayer(commands[1]);
                }
            }
            default: {}
        }
        diff := time.Now().Sub(startTime).Milliseconds();
        if diff < 100-overflow {
            continue;
        }
        // For all idle players
        temp := m.players;
        for temp != nil {
            select {
                case pl_message := <-temp.player.ws.channel: {
                    if pl_message ==  "" || temp.player.ws.isClosed() {
                        fmt.Println("dead connection", temp, temp.player);
                        delete(m.playing, temp.player.id);
                        m.players.remove(&m.players,temp.player.id);
                        break;
                    }
                    cmds := strings.Split(pl_message, " ")
                    if len(cmds) < 2 {break;}
                    switch cmds[0] {
                    //TODO: Somwhow have gametype here
                    case "find": {
                        m.players.remove(&m.players,temp.player.id)
                        //TODO: check for the correct type
                        temp.gameType = cmds[1];
                        if m.finders == nil {
                            m.finders = temp;
                        } else {
                            m.finders.add(&m.players,temp)
                        }
                    }
                }
            }
            default: {
                temp.finding_tick++;
            }
            }
            if temp.finding_tick > 10*3 {
                temp.player.ws.Close();
                delete(m.playing, temp.player.id);
                m.players.remove(&m.players,temp.player.id);
                m.finders.remove(&m.players,temp.player.id);
            }
            temp = temp.next;
        }
        // for all finding players
        temp = m.finders;
        for temp != nil {
                if temp.player.ws.closed {
                    m.players.remove(&m.players,temp.player.id);
                    temp = temp.next;
                    continue;
                }
                select {
                case pl_message:= <- temp.player.ws.channel: {
                    switch pl_message {
                        case "": {
                            delete(m.playing, temp.player.id);
                            m.finders.remove(&m.players,temp.player.id);
                        }
                        case "stop": {
                            m.finders.remove(&m.players,temp.player.id);
                            if m.finders == nil {
                                m.players = temp;
                            } else {m.players.add(&m.players,temp)}
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
    return;
}
