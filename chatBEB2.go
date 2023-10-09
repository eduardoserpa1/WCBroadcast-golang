package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"strconv"
	"sort"

	. "SD/BEB"
)

type WaitingCausalBroadcast struct {
	PendingMessages map[string][]BestEffortBroadcast_Ind_Message
	VectorClock     *VectorClock
}

func NewWaitingCausalBroadcast(processes []string) *WaitingCausalBroadcast {
	return &WaitingCausalBroadcast{
		PendingMessages: make(map[string][]BestEffortBroadcast_Ind_Message),
		VectorClock:     NewVectorClock(processes),
	}
}

func caused(c1 *VectorClock, c2 map[string]int) bool {
	for process, timestamp := range c1.Clock {
		if c2[process] >= timestamp {
			return false
		}
	}
	return true
}

func (wcb *WaitingCausalBroadcast) Broadcast(message string, sender string) {
	wcb.VectorClock.Increment(sender)

	req := BestEffortBroadcast_Req_Message{
		Addresses:    []string{sender},
		Message:      message + "§" + sender,
		VectorClock:  wcb.VectorClock.Clock,
	}

	for _, receiver := range wcb.VectorClock.Clock {
		if receiver != sender {
			wcb.PendingMessages[receiver] = append(wcb.PendingMessages[receiver], BestEffortBroadcast_Ind_Message{
				Message:      req.Message,
				VectorClock:  req.VectorClock,
				From:         sender,
				Addresses:    []string{receiver},
			})
		}
	}

	wcb.DeliverMessages()
}

func (wcb *WaitingCausalBroadcast) DeliverMessages() {
	for {
		for receiver, messages := range wcb.PendingMessages {
			deliverableMessages := []BestEffortBroadcast_Ind_Message{}

			for _, message := range messages {
				if caused(wcb.VectorClock, message.VectorClock) {
					message.Deliverable = true
					wcb.VectorClock.Increment(message.From)
					fmt.Printf("Entregando mensagem de %s: %s\n", message.From, message.Message)
				} else {
					deliverableMessages = append(deliverableMessages, message)
				}
			}

			wcb.PendingMessages[receiver] = deliverableMessages
		}

		// Aguarde um curto período de tempo antes de verificar novamente
		time.Sleep(time.Millisecond * 100)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please specify at least one address:port!")
		fmt.Println("go run chatBEB.go 127.0.0.1:5001 127.0.0.1:6001 127.0.0.1:7001")
		return
	}

	addresses := os.Args[1:]
	fmt.Println(addresses)

	wcb := NewWaitingCausalBroadcast(addresses)

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		var msg string

		for {
			if scanner.Scan() {
				msg = scanner.Text()
				wcb.Broadcast(msg, addresses[0])
			}
		}
	}()

	blq := make(chan int)
	<-blq
}
