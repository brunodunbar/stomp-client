package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"./stompclient"
)

func main() {
	c, err := stompclient.Create("tcp", "localhost:61616", "pessoal-user", "pessoal-user")

	if err != nil {
		fmt.Println(err)
		return
	}

	c.Subscribe("jms.queue.test", func(message []byte) (bool, error) {
		fmt.Println(string(message))
		return true, nil
	})

	c.Send("jms.queue.test", "application/json", []byte("{\"status\": \"OK\"}"))

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	_ = <-sigs
	c.Close()
}