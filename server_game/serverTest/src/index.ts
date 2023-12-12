import WebSocket from "ws"
import axios from "axios"
const gameAmount: number = 3;
interface Player{
    id: String;
    authToken: String;
    rating: number;
    ws?: WebSocket;
}
const players = new Map<String, Player>();
async function main() {
    for (let i = 0; i < gameAmount;i++ ){
        const data = await axios.post("http://localhost:4000/login", {id: "asdf"+i, name: "name_"+i, rating: i*10});
        const jsondata:Player = data.data as Player
        const player: Player= {id:jsondata.id , authToken:jsondata.authToken, rating: jsondata.rating}
        const ws = new WebSocket("ws://localhost:3000/game",{
            headers: {
            "X-Auth-Cb" : ""+player.authToken 
        }});
        ws.on("connect", ()=>{player.ws = ws; players.set(player.id, player)});
        ws.on('error', ()=>{console.log(player.id,"error ws")})
        ws.on('close', ()=>{players.delete(player.id) ;console.log(player.id, "connection closed")})
    }
}
main();
