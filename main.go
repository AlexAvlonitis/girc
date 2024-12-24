package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

func userInput(ch chan string) {
	for {
		fmt.Print("Enter command> ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		input := scanner.Text()

		ch <- input
	}
}

func main() {
	client := NewClient("halcyon.il.us.dal.net", 6669, "CCClient", "CCClient", "CCClient")
	err := client.Connect()
	if err != nil {
		log.Fatalf("Error connecting to server: %s", err)
		return
	}
	ch := make(chan []byte)
	go client.Read(ch)

	userInputCh := make(chan string)
	go userInput(userInputCh)

	for {
		select {
		case msg := <-ch:
			fmt.Printf("%s\n", msg)
		case userInput := <-userInputCh:
			command := ParseCommand(userInput)
			client.Write(command)
		}
	}
}
