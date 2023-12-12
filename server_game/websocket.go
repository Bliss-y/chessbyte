package main
import (
    "fmt"
    "net"
    "crypto/sha1"
    "bufio"
    "net/http"
    "io"
    "bytes"
    "encoding/base64"
)

type WebSocketConnection struct {
    connection net.Conn;
    closed bool;
    channel chan string;
    isReading bool;
    authToken string;
}

type withConnection interface{
    getConnection() net.Conn;
    isClosed() bool;
    getChannel() chan string;
    Closed();
}

func (w WebSocketConnection) getConnection() net.Conn{
    return w.connection;
}

func (w *WebSocketConnection) isClosed() bool {
    return w.closed;
}


func (ws*WebSocketConnection) sendWebscoketMessage( message string) bool {
    if ws.isClosed() {
        return false;
    }
    var responsebytes bytes.Buffer;
    responsebytes.WriteByte(uint8(129));
    responsebytes.WriteByte(uint8(len(message)));
    responsebytes.WriteString(message);
    _,err:=ws.getConnection().Write(responsebytes.Bytes());
    if err != nil {
        return false;
    } else {
        return true;
    }
}

func (ws *WebSocketConnection) getChannel() chan string{
    return ws.channel;
}

func (ws *WebSocketConnection) setReading(x bool) {
    ws.isReading = x;
}

func (conn *WebSocketConnection) Close() {
    fmt.Println("Connection closed", conn);
    if(conn.closed){return}
    close(conn.channel);
    conn.connection.Close();
    conn.closed = true;
}

func (connectionHolder *WebSocketConnection) read() {
    defer connectionHolder.setReading(false);
    conn := connectionHolder.getConnection();
    connectionHolder.isReading = true;
    message := connectionHolder.getChannel();
    buffer := make([]byte, 50); 
    for !connectionHolder.isClosed(){
        _, err := conn.Read(buffer);
        if err == io.EOF {
            message <- "Connection Closed!";
            fmt.Println("Connection Closed!");
            connectionHolder.Close();
            break;
        } else if err != nil {
          fmt.Println("Err Occurred.. ?", err);
          connectionHolder.Close();
          break;
        }
        // do the message parsing
        fin := 0;
        fin = int((buffer[0] >> 7) & 1)
        if fin != 1 {
            fmt.Println("fin is not 1")
        }
        opcode :=0;
        opcode = int(buffer[0] & 0x0F);
        if opcode != 0x1 {
            fmt.Println("opcode is not 1")
        }
        mask := (buffer[1] >> 7) & 1;
        if mask != 1 {
            fmt.Println("masbit is not 1")
        }
        payload_length := buffer[1] & 0x7F; // last 7 bits
        current := 2;
        if (payload_length == 126) {
            // read next 16 bytes for payload lengthh
            payload_length = buffer[2];
            current = 4;
        } else if (payload_length == 127) {
            // read next 64 bytes for payload length
            payload_length = buffer[2];
            current = 10;
        }
        masking_key := []byte{buffer[current + 0], buffer[current + 1], buffer[current + 2], buffer[current + 3]};
        current += 4;
        msg := make([]byte, payload_length);
        for i := 0; i < int(payload_length); i++ {
            msg[i] = buffer[i + current] ^ masking_key[i % 4];
        }
        message <- string(msg)
    }
    fmt.Println("reading ended", connectionHolder);
}

func closeSocket(conn net.Conn) bool{
    var responsebytes bytes.Buffer;
    responsebytes.WriteByte(uint8(0x87));
    responsebytes.WriteByte(uint8(0));
    _,err:=conn.Write(responsebytes.Bytes());
    conn.Close();
    return err != nil;
}

func handleRequest(conn net.Conn, onConnection func(conn *WebSocketConnection)){
        req,err := http.ReadRequest(bufio.NewReader(conn))
        if err != nil{
            conn.Close();
            return
        }
        if req.URL.Path == "/game" {
            // do the websocket handshake
            if req.Method != "GET" {
                conn.Close();   
                return;
            }
            connection_header, exists:= req.Header["Upgrade"];
            if !exists || connection_header[0] != "websocket" {
                conn.Close();   
                return;
            }
            websocketkey, exists := req.Header["Sec-Websocket-Key"];
            websocketkey[0] += "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
            hash := sha1.New()
	        // Write the string data to the hash object
            hash.Write([]byte(websocketkey[0]));
	        // Get the final hash as a byte slice
	        hashBytes:= hash.Sum(nil)
            encoded:= base64.StdEncoding.EncodeToString(hashBytes);    
            var responsebytes bytes.Buffer;
            fmt.Fprintf(&responsebytes, "HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: %s\r\n\r\n", encoded);
            _, err := conn.Write([]byte(responsebytes.String()));
            if err != nil {
                conn.Close();
                return
            }
            // check for existing games first
            connection_header, exists = req.Header["X-Auth-Cb"];
            if !exists {
                conn.Close();
                return;
            }
            // TODO: JWT ?
            ws := &WebSocketConnection{conn, false,make(chan string), false, connection_header[0]}
            go ws.read();
            onConnection(ws);
            return
        }
        conn.Close();
}
