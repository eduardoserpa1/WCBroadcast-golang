// Construido como parte da disciplina: Sistemas Distribuidos - PUCRS - Escola Politecnica
//  Professor: Fernando Dotti  (https://fldotti.github.io/)

/*
LANCAR N PROCESSOS EM SHELL's DIFERENTES, UMA PARA CADA PROCESSO, O SEU PROPRIO ENDERECO EE O PRIMEIRO DA LISTA
go run chatBEB.go 127.0.0.1:5001  127.0.0.1:6001  127.0.0.1:7001   // o processo na porta 5001
go run chatBEB.go 127.0.0.1:6001  127.0.0.1:5001  127.0.0.1:7001   // o processo na porta 6001
go run chatBEB.go 127.0.0.1:7001  127.0.0.1:6001  127.0.0.1:5001     // o processo na porta ...
*/

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"strconv"
	"sort"
	"encoding/json"
	. "SD/BEB"
)


type VectorClock struct {
	Clock map[string]int
}

func NewVectorClock(processes []string) *VectorClock {
	clock := make(map[string]int)
	for _, process := range processes {
		clock[process] = 0
	}
	return &VectorClock{Clock: clock}
}

func (vc *VectorClock) Increment(processID string) {
	vc.Clock[processID]++
}

func formatVectorClock(clock map[string]int) string {
	// Extract keys from the map
	var keys []string
	for key := range clock {
		keys = append(keys, key)
	}

	// Sort keys in ascending order
	sort.Strings(keys)

	// Create the formatted string
	formattedClock := "["
	for _, key := range keys {
		value := clock[key]
		formattedClock += key[10:] + " : " + strconv.Itoa(value) + ", "
	}
	formattedClock = strings.TrimSuffix(formattedClock, ", ")
	formattedClock += "]"
	return formattedClock
}

func caused(c1 *VectorClock, c2 map[string]int) bool {
	at_least_once := 0 
	for process, timestamp := range c1.Clock {
		if timestamp > c2[process] {
			return false
		}
		if timestamp < c2[process] {
			at_least_once++
		}
	}
	if at_least_once > 0{
		return true
	}
	return false
}


func dumpmap(vc map[string]int) {
	fmt.Println("Vector Clock:")
	for key, value := range vc {
		fmt.Printf("%s: %d\n", key[10:], value)
	}
}

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Please specify at least one address:port!")
		fmt.Println("go run chatBEB.go 127.0.0.1:5001  127.0.0.1:6001   127.0.0.1:7001")
		fmt.Println("go run chatBEB.go 127.0.0.1:6001  127.0.0.1:5001   127.0.0.1:7001")
		fmt.Println("go run chatBEB.go 127.0.0.1:7001  127.0.0.1:6001   127.0.0.1:5001")
		return
	}


	//myadress := os.Args[0]
	addresses := os.Args[1:]
	fmt.Println(addresses)

	beb := BestEffortBroadcast_Module{
		Req: make(chan BestEffortBroadcast_Req_Message),
		Ind: make(chan BestEffortBroadcast_Ind_Message)}

	//beb.Init(addresses[0])
	beb.InitD(addresses[0], false)

	// enviador de broadcasts
	// Inicializa o vetor de relógio vetorial
	vectorClock := NewVectorClock(addresses)

	// Enviador de broadcasts
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		var msg string

		for {
			if scanner.Scan() {
				msg = scanner.Text()
				msg += "§" + addresses[0]
			}

			vectorClock.Increment(addresses[0])
		
			req := BestEffortBroadcast_Req_Message{
				Addresses: addresses[0:],
				Message:   msg,
				VectorClock: vectorClock.Clock,
			}

			fmt.Printf("--------\nreq:\n%v\n---------\n", req)


			beb.Req <- req // ENVIA PARA TODOS OS PROCESSOS ENDEREÇADOS NO INÍCIO
		}
	}()

	// Receptor de broadcasts
	go func() {
		for {
			in := <-beb.Ind // RECEBE MENSAGEM DE QUALQUER PROCESSO
			message := strings.Split(in.Message, "§")
			in.From = message[1]
			in.Message = message[0]

			dumpmap(vectorClock.Clock)
			fmt.Println(in.VectorClock)

			if caused(vectorClock, in.VectorClock){
				fmt.Printf("%v CAUSOU %v \n",message,in.Message)
			}

			// Atualiza o relógio vetorial com o relógio da mensagem recebida
			for processID, value := range in.VectorClock {
				if vectorClock.Clock[processID] < value {
					vectorClock.Clock[processID] = value
				}
			}

			// Incrementa o relógio vetorial do processo que enviou a mensagem
			vectorClock.Increment(in.From)

			formattedClock := formatVectorClock(vectorClock.Clock)
        	fmt.Printf("Message from %v: %v (Vector Clock: %v)\n", in.From, in.Message, formattedClock)
 
		}
	}()

	blq := make(chan int)
	<-blq
}
