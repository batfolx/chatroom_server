package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

func main() {

	host := "localhost:10000"
	proto := "tcp4"
	StartServer(host, proto)

}

/* This struct holds the messages and users connected to a specific chat room */
type ChatRoomData struct {
	messages    []string
	connections []*ConnInfo
}

/* This struct keeps a connection object along with the persons name and the chat room they are currently in */
type ConnInfo struct {
	connection      *net.Conn
	name            string
	currentChatRoom string
}

// we will have map of strings to ChatRoomData structs
// the map will have a key of a chat room name (chat room 1)
// ChatRoomData will contain the messages of the chatroom, with the connections inside of it
var chatRooms map[string]*ChatRoomData

func StartServer(host string, protocol string) {
	fmt.Println("Starting server")

	//connections = map[string]*ConnInfo{}
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
			buffer := make([]byte, 1028)
			n, err := conn.Read(buffer)
			if err != nil {
				fmt.Printf("Did not get name from user")
				conn.Close()
			} else {
				name := string(buffer[:n])

				fmt.Println("Got connection from " + name)
				// once we get the name, we prompt the user to get the chat room they want to go to
				chatRoom := chooseChatRoom(&conn)

				// create a ConnInfo struct to store information about a particular connection
				connInfo := &ConnInfo{
					connection:      &conn,
					name:            name,
					currentChatRoom: chatRoom,
				}

				fmt.Printf("Chatroom number: %s\n", chatRoom)
				addUserToChatRoom(chatRoom, connInfo)
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

	/* We send the messages to the user asynchronously */
	go func() {
		for _, msg := range chatRooms[chatRoom].messages {
			(*conn).Write([]byte(msg + "\n"))
		}
	}()

	for {
		buffer := make([]byte, 512)
		n, err := (*conn).Read(buffer)
		if err != nil {
			fmt.Printf("Lost connection from %s\n", (*conn).RemoteAddr())
			return
		} else {
			if checkCommands(conn, string(buffer[0:n]), &chatRoom) {
				continue
			}
			/* read in the message, then add it to all of the messages */
			message := []byte(fmt.Sprintf("%s: %s", name, string(buffer[:n])))
			chatRooms[chatRoom].messages = append(chatRooms[chatRoom].messages, string(message))
			for i, conn := range chatRooms[chatRoom].connections {
				if conn == nil {
					continue
				}
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
	/* If the chat room exists already, we can append messages and connections directly */
	if msgs, exists := chatRooms[chatRoom]; exists {
		conns = msgs.connections
		conns = append(conns, conn)
		chatRooms[chatRoom].connections = conns
	} else {
		// else we need to create the chat room
		conns = append(conns, conn)
		data := ChatRoomData{
			messages:    messages,
			connections: conns,
		}
		chatRooms[chatRoom] = &data
	}
}

func checkCommands(conn *net.Conn, buffer string, chatRoom *string) bool {

	if buffer == "users all" {
		users := ""

		for _, chatRoomData := range chatRooms {
			for _, c := range chatRoomData.connections {
				if c == nil {
					continue
				}
				users += fmt.Sprintf("User: %s; Online\n", (*c).name)
			}
		}
		(*conn).Write([]byte(users))
		return true
	}
	if buffer == "users" {
		users := ""
		for _, c := range chatRooms[*chatRoom].connections {
			if c == nil {
				continue
			}
			users += fmt.Sprintf("User: %s; Online\n", (*c).name)
		}
		(*conn).Write([]byte(users))
		return true
	}
	if strings.Contains(buffer, "switch") {
		newChatRoom := strings.Fields(buffer)[1]
		fmt.Printf("New chat room: %s\n", newChatRoom)
		switchChatRooms(conn, chatRoom, newChatRoom)
		(*conn).Write([]byte(fmt.Sprintf("Switched from chat room %s to chat room %s!", chatRoom, newChatRoom)))
		return true
	}
	return false
}

func switchChatRooms(conn *net.Conn, oldChatRoom *string, newChatRoom string) {
	// TODO fix switching a user into a different chat room
	tempConn := ConnInfo{}
	index := -1
	for i, c := range chatRooms[*oldChatRoom].connections {
		if *conn == nil || *c.connection == nil {
			continue
		}
		if *conn == *c.connection {
			tempConn.connection = conn
			tempConn.currentChatRoom = *oldChatRoom
			tempConn.name = (*c).name
			index = i
			break
		}
	}

	addUserToChatRoom(newChatRoom, &tempConn)
	if index != -1 {
		chatRooms[*oldChatRoom].connections[index] = nil
		*oldChatRoom = newChatRoom
	}

	// we change the value of the chatroom to be pointing to something different

}
