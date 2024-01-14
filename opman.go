package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"golang.zx2c4.com/irc/hbot"
)

type Config struct {
	Channels []string `json:"channels"`
	Nick     string   `json:"nick"`
	Host     string   `json:"host"`
	OpNicks  []string `json:"opNicks"`
}

func getConfig(fileName string) Config {
	fileStream, err := os.Open(fileName)
	if err != nil {
		log.Fatal("Can't open file!")
	}
	bytes, err := io.ReadAll(fileStream)
	if err != nil {
		log.Fatal("Error reading file")
	}
	var config Config

	err = json.Unmarshal(bytes, &config)
	if err != nil {
		log.Fatal("Error parsing file")
	}

	return config
}

func main() {
	config := getConfig("go-opman.json")

	botConfig := hbot.Config{
		Host:     config.Host,
		Nick:     config.Nick,
		Realname: fmt.Sprintf("Mr. %s", config.Nick),
		User:     strings.ToLower(config.Nick),
		Channels: config.Channels,
		Logger:   hbot.Logger{Verbosef: log.Printf, Errorf: log.Printf},
	}

	bot := hbot.NewBot(&botConfig)

	bot.AddTrigger(hbot.Trigger{
		Condition: func(b *hbot.Bot, m *hbot.Message) bool {
			if m.Command != "JOIN" {
				return false
			}
			return true
		},
		Action: func(b *hbot.Bot, m *hbot.Message) {
			chn := m.Param(0)
			for _, opNick := range config.OpNicks {
				if m.Prefix.Name == opNick {
					fmt.Printf("Giving op to %s in %s\n", m.Prefix.Name, chn)
					b.ChMode(m.Prefix.Name, chn, "+o")
				}
			}
		},
	})

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for range c {
			bot.Close()
			os.Exit(0)
		}
	}()

	for {
		bot.Run()
		time.Sleep(time.Second * 5)
	}
}
