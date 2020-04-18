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
	connections []*net.Conn
}

var connections map[string]*net.Conn

// we will have map of strings to ChatRoomData structs
// the map will have a key of a chat room name (chat room 1)
// ChatRoomData will contain the messages of the chatroom, with the connections inside of it
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
				chatRoom := chooseChatRoom(&conn)
				fmt.Printf("Chatroom number: %s\n", chatRoom)
				go handleConnection(&conn, name)
			}
		}
	}
}

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

func handleChatRoom(chatRoom string, conn *net.Conn, message string) {
	var messages []string
	var conns []*net.Conn

	/* If the chat room exists already, we can append messages and connections directly */
	if msgs, exists := chatRooms[chatRoom]; exists {

		// get the existing messages from the ChatRoomData struct
		// and add the new message to the list of messages and send it to everyone
		messages = msgs.messages
		messages = append(messages, message)
		conns = msgs.connections
		conns = append(conns, conn)

		chatRooms[chatRoom].connections = conns
		chatRooms[chatRoom].messages = messages
		for _, c := range chatRooms[chatRoom].connections {
			(*c).Write([]byte(message))
		}
	} else {
		messages = append(messages, message)
		conns = append(conns, conn)
		data := ChatRoomData{
			messages:    messages,
			connections: conns,
		}
		chatRooms[chatRoom] = &data
	}
}
