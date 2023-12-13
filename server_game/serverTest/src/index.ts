import WebSocket from "ws"
import axios from "axios"
const gameAmount: number = 100000;
interface Player{
    id: String;
    authToken: String;
    rating: number;
    ws?: WebSocket;
}
const players = new Map<String, Player>();
let ec = 0;
async function main() {
    for (let i = 0; i < gameAmount;i++ ){
        const data = await axios.post("http://localhost:4000/login", {id: "asdf"+i, name: "name_"+i, rating: i*10});
        const jsondata:Player = data.data as Player
        const player: Player= {id:jsondata.id , authToken:jsondata.authToken, rating: jsondata.rating}
        const ws = new WebSocket("ws://localhost:3000/game",{
            headers: {
            "X-Auth-Cb" : ""+player.authToken 
        }});
        ws.on("open", ()=>{player.ws = ws; players.set(player.id, player); console.log("connection opened", player.id)});
        //ws.on('error', (e)=>{console.log(player.id,"error ws"); console.log(++ec)})
        //ws.on('close', ()=>{players.delete(player.id);})
    }
}
main()
