package main

import (
	"fmt"
	"io"
	"log"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatal("Connection error:", err)
	}
	defer conn.Close()

	response, err := io.ReadAll(conn)
	if err != nil {
		log.Fatal("Read error:", err)
	}

	if string(response) != "OK\n" {
		log.Fatalf("Unexpected response: %q", string(response))
	}

	fmt.Println("Server response:", string(response))
	fmt.Println("Response is correct!")
}

