package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	host := "localhost:10000"
	proto := "tcp4"
	connection := connectToServer(host, proto)
	go getMessages(connection)
	chat(connection)
}

func connectToServer(host string, proto string) *net.Conn {
	for {
		conn, err := net.Dial(proto, host)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			continue
		} else {
			return &conn
		}
	}

}

func chat(conn *net.Conn) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Please enter your name: ")
	for {
		msg, err := reader.ReadString('\n')
		msg = strings.TrimSuffix(msg, "\n")
		if err != nil {
			(*conn).Close()
			fmt.Printf("Error: %s\n", err)
			return
		} else {
			(*conn).Write([]byte(msg))
		}
	}
}

func getMessages(conn *net.Conn) {
	for {
		buffer := make([]byte, 512)
		n, err := (*conn).Read(buffer)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			(*conn).Close()
			return
		} else {
			fmt.Println(string(buffer[0:n]))
		}
	}

}
