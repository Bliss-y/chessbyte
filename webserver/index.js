import WebSocket, { WebSocketServer } from 'ws';

var gameserver = new WebSocket("ws://localhost:3000/game", {
headers: {
    "x-auth-cb": 'SERVER_TOKEN_FORNOW'
}
} );
gameserver.on('open', function(){
    console.log("connection successfully opened")
})

gameserver.on("close", function(){
    console.log("connection closed");
})
gameserver.on("error", function(err){
    console.log(err)
})

