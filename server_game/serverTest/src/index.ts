import WebSocket from "ws"
const gameAmount: number = 300;
interface Player{
    id: String;
    authToken: String;
    ws?: WebSocket;
}
const players = new Map<String, Player>();
async function main() {
    for (let i = 0; i < gameAmount;i++ ){
        const data = await fetch("http://localhost/loginUser", {method: "POST",body: JSON.stringify({id:""+i, name: "name_"+i }) });
        const jsondata:Player = await data.json() as Player
        const player: Player= {id:jsondata.id , authToken:jsondata.authToken   }
        const ws = new WebSocket("ws://localhost:3000",[], {
            headers: {
            "X-Auth-Cb" : ""+player.authToken 
        }});
        ws.on("connect", ()=>{player.ws = ws; players.set(player.id, player)});
    }
}
main();
