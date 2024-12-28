package main

import (
	"bufio"
	"fmt"
	"girc/commands"
	"girc/connection"
	"log"
	"os"
)

func main() {
	client := connection.NewClient("halcyon.il.us.dal.net", 6669, "CCClient", "CCClient", "CCClient")
	err := client.Connect()
	if err != nil {
		log.Fatalf("Error connecting to server: %s", err)
		return
	}

	// Create a done channel, where we will send a signal to stop the program gracefully
	done := make(chan interface{})
	defer close(done)

	// Create a channel to read from the connection
	ch := make(chan []byte)
	go client.Read(ch, done)

	// Create a channel to read user input
	userInputCh := make(chan string)
	go userInput(userInputCh, done)

	for {
		select {
		case <-done:
			return
		case msg := <-ch:
			fmt.Printf("%s\n", msg)
			fmt.Print("Enter command> ")
		case userInput := <-userInputCh:
			commands.SendCommand(userInput, client)
			fmt.Print("Enter command> ")
		}
	}
}

func userInput(ch chan string, done chan interface{}) {
	for {
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		input := scanner.Text()

		select {
		case <-done:
			return
		default:
			ch <- input
		}
	}
}
