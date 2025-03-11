package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {return true},
	}
	scaleIP string 
	scalePort string 
)

func init() {
	err := godotenv.Load("../.env")
	if err !=nil {
		log.Fatalf("Error loading .env file") 
	}
	scaleIP = os.Getenv("SCALE_IP_ADDR")
	scalePort = os.Getenv("SCALE_PORT")
}

func connectToScale () (net.Conn, error) {
	address := net.JoinHostPort(scaleIP, scalePort)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to scale: %v", err)
	}
	return conn, nil 
}

func websocketHandler(w http.ResponseWriter, r *http.Request) {
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade failed: ", err)
		return 
	}
	defer wsConn.Close()

	scaleConn, err := connectToScale()
	if err != nil {
		log.Println(err)
		return
	}
	defer scaleConn.Close()

	go func() {
		for {
			_, message, err := wsConn.ReadMessage()
			if err != nil {
				log.Println("Websocket read error: ", err)
				break
			}
			log.Printf("Websocket recevied: %s", message)
			if strings.HasPrefix(string(message), "cmd:") {
				command := strings.TrimPrefix(string(message), "cmd:")
				output, err := runCommand(command) 
				if err != nil {
					log.Println("Command exec err: ", err)
					wsConn.WriteMessage(websocket.TextMessage, []byte("Command exec error: "+err.Error()))
					continue 
				}
				wsConn.WriteMessage(websocket.TextMessage, output)
				continue
			}
			if _, err := scaleConn.Write(message); err != nil {
				log.Println("Scale write error: ", err)
				break
			}
		}
	} ()
	buf := make([]byte, 1024)
	for {
		n, err := scaleConn.Read(buf)
		if err != nil {
			log.Println("Scale read error: ", err)
			break
		}
		log.Printf("Scale received: %s", buf[:n])
		if err := wsConn.WriteMessage(websocket.TextMessage, buf[:n]); err != nil {
			log.Println("Websocket write error: ", err)
			break 
		}
	}
}

func runCommand(command string) ([]byte, error) {
    var cmd *exec.Cmd
    if os.PathSeparator == '\\' { // Windows system
        cmd = exec.Command("cmd.exe", "/C", command)
    } else { // Unix-like system
        cmd = exec.Command("sh", "-c", command)
    }
    return cmd.CombinedOutput()
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hi")
}

func main() {
	http.HandleFunc("/", helloHandler)
	http.HandleFunc("/websocket", websocketHandler)

	port := ":3000"
	fmt.Printf("Server running on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

