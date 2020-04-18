package main

import (
	"fmt"
	"net"
)

func main() {

	host := "192.168.1.7:10000"
	proto := "tcp4"
	StartServer(host, proto)

}

type ChatRoomData struct {
	messages []string
	connections []*net.Conn
}

var connections map[string]*net.Conn
var chatRooms map[string]*ChatRoomData

func StartServer(host string, protocol string) {
	fmt.Println("Starting server")

	connections = map[string]*net.Conn{}

	listener, err := net.Listen(protocol, host)
	if err != nil {
		fmt.Println("Could not start server on " + host)
		fmt.Printf("Error %s\n", err)
		return
	}
	fmt.Println("Server successfully established!")
	for {

		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error in establishing connectiont to %s\n", conn)
		} else {
			fmt.Printf("Connections: %v\n", connections)
			buffer := make([]byte, 2048)
			n, err := conn.Read(buffer)
			if err != nil {
				fmt.Printf("Did not get name from user")
			} else {
				name := string(buffer[:n])
				connections[name] = &conn
				fmt.Println("Got connection from " + name)

				go handleConnection(&conn, name)
			}

		}

	}

}

func handleConnection(conn *net.Conn, name string) {
	defer (*conn).Close()
	for {

		buffer := make([]byte, 512)
		n, err := (*conn).Read(buffer)
		if err != nil {
			fmt.Printf("Lost connection from %s\n", (*conn).RemoteAddr())
			return
		} else {
			if string(buffer[0:n]) == "users" {
				users := ""

				for name, _ := range connections {
					users += fmt.Sprintf("User: %s; Online\n", name)
				}
				(*conn).Write([]byte(users))
				continue
			}
			message := fmt.Sprintf("%s: %s", name, string(buffer[:n]))
			for key, connection := range connections {
				_, err = (*connection).Write([]byte(message))
				if err != nil {
					delete(connections, key)
				}

			}
		}

	}

}
func handleChatRoom(chatRoom string, conn *net.Conn) {
	var messages []string
	var conns []*net.Conn
	if msgs, exists := chatRooms[chatRoom]; exists {
		messages = msgs.messages
		conns = msgs.connections
		conns = append(conns, conn)
	} else {




		_ = ChatRoomData {
			messages:    messages,
			connections: conns,
		}



	}


}