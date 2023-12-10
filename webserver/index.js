import WebSocket, { WebSocketServer } from 'ws';
import {http} from "http"

/**
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
    **/
const proxy = http.createServer((req, res) => {
  res.writeHead(200, { 'Content-Type': 'text/plain' });
  res.end('okay');
});

proxy.on('connect', (req, clientSocket, head) => {
  // Connect to an origin server
  const { port, hostname } = new URL(`http://${req.url}`);
  const serverSocket = net.connect(port || 80, hostname, () => {
    clientSocket.write('HTTP/1.1 200 Connection Established\r\n' +
                    'Proxy-agent: Node.js-Proxy\r\n' +
                    '\r\n');
      serverSocket.write(JSON.stringify({id: 100, rating: 200, auth: true, name: "Player"}));
    serverSocket.pipe(clientSocket);
    clientSocket.pipe(serverSocket);
  });
});

proxy.listen(4000, '127.0.0.1', () => {

  // Make a request to a tunneling proxy
    console.log("server running")
})
