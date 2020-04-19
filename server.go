package main

import (
	"fmt"
	"net"
	"strconv"
)

func main() {

	host := "localhost:10000"
	proto := "tcp4"
	StartServer(host, proto)

}

type ChatRoomData struct {
	messages    []string
	connections []*ConnInfo
}

type ConnInfo struct {
	connection *net.Conn
	name       string
}

var connections map[string]*ConnInfo

// we will have map of strings to ChatRoomData structs
// the map will have a key of a chat room name (chat room 1)
// ChatRoomData will contain the messages of the chatroom, with the connections inside of it
var chatRooms map[string]*ChatRoomData

func StartServer(host string, protocol string) {
	fmt.Println("Starting server")

	connections = map[string]*ConnInfo{}
	chatRooms = map[string]*ChatRoomData{}

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
				connections[name] = &ConnInfo{
					connection: &conn,
					name:       name,
				}
				fmt.Println("Got connection from " + name)
				chatRoom := chooseChatRoom(&conn)
				fmt.Printf("Chatroom number: %s\n", chatRoom)
				addUserToChatRoom(chatRoom, connections[name])
				go handleConnection(&conn, name, chatRoom)
			}
		}
	}
}

/* We get the chat room that the user wants to select and return it to index */
func chooseChatRoom(conn *net.Conn) string {

	// we send the options to the user
	options := `
		Please choose the chat room that you would like to go into.
		Just enter a number and you'll go into that chat room
	`
	_, err := (*conn).Write([]byte(options))

	// if there is an error, probably means the connection was interrupted, we return -1
	if err != nil {
		fmt.Printf("Error! %s\n", err)
		return "-1"
	}

	// we loop until the user enters a number
	for {

		buffer := make([]byte, 128)
		n, err := (*conn).Read(buffer)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			(*conn).Close()
			return "-1"
		}

		// we check if we can convert the number, if err is nil then we converted
		if _, err := strconv.Atoi(string(buffer[0:n])); err == nil {
			return string(buffer[0:n])
		} else {
			(*conn).Write([]byte("You need to specify a single number."))
		}
	}
}

/* This handles the sending of the messages  */
func handleConnection(conn *net.Conn, name string, chatRoom string) {
	defer (*conn).Close()
	for {
		buffer := make([]byte, 512)
		n, err := (*conn).Read(buffer)
		if err != nil {
			fmt.Printf("Lost connection from %s\n", (*conn).RemoteAddr())
			return
		} else {
			if string(buffer[0:n]) == "users all" {
				users := ""

				for name, _ := range connections {
					users += fmt.Sprintf("User: %s; Online\n", name)
				}
				(*conn).Write([]byte(users))
				continue
			}
			if string(buffer[0:n]) == "users" {
				users := ""
				for _, c := range chatRooms[chatRoom].connections {
					users += fmt.Sprintf("User: %s; Online\n", (*c).name)
				}
				continue
			}

			message := []byte(fmt.Sprintf("%s: %s", name, string(buffer[:n])))
			for i, conn := range chatRooms[chatRoom].connections {
				_, err := (*conn.connection).Write(message)
				if err != nil {
					chatRooms[chatRoom].connections = append(chatRooms[chatRoom].connections[i:], chatRooms[chatRoom].connections[:i+1]...)
				}
			}
		}
	}
}

func addUserToChatRoom(chatRoom string, conn *ConnInfo) {
	var messages []string
	var conns []*ConnInfo
	fmt.Printf("Before checking if key exists: %v\n", chatRooms[chatRoom])
	/* If the chat room exists already, we can append messages and connections directly */
	if msgs, exists := chatRooms[chatRoom]; exists {
		conns = msgs.connections
		conns = append(conns, conn)
		chatRooms[chatRoom].connections = conns
	} else {
		conns = append(conns, conn)
		data := ChatRoomData{
			messages:    messages,
			connections: conns,
		}
		chatRooms[chatRoom] = &data
	}
}

func checkCommands(buffer string, conns map[string]*net.Conn) bool {

	if buffer == "users" {

	}
	return true

}

func switchChatRooms(conn *net.Conn, chatRoom string) {}
