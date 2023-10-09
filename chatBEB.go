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
	"strconv"
	"strings"

	. "SD/BEB"
)

func inc(c [10][2]string, target string, q int) [10][2]string {
	for i := 0; i < len(c); i++ {
		if c[i][0] == target {
			intc, err := strconv.Atoi(c[i][1])
			if err != nil {
				fmt.Println("Erro:", err)
				return c
			}
			intc += q
			c[i][1] = strconv.Itoa(intc)
		}
	}
	return c
}

func equal(c [10][2]string, target string, q int) [10][2]string {
	for i := 0; i < len(c); i++ {
		if c[i][0] == target {
			intc, err := strconv.Atoi(c[i][1])
			if err != nil {
				fmt.Println("Erro:", err)
				return c
			}
			intc = q
			c[i][1] = strconv.Itoa(intc)
		}
	}
	return c
}
func stringToClock(s string) [10][2]string {
	var clock [10][2]string
	splitted := strings.Split(s, ",")
	for i := 0; i < 10; i++ {
		if i < len(splitted) {
			spl := strings.Split(splitted[i], "&")
			clock[i][0] = spl[0]
			clock[i][1] = spl[1]
		}
	}
	return clock
}

func clockToString(c [10][2]string) string {
	var r string = ""
	for i := 0; i < 10; i++ {
		if c[i][0] != "" && c[i][0] != " " {
			r += c[i][0] + "&" + c[i][1] + ","
		}
	}
	return r[:len(r)-1]
}

func formatClock(clock [10][2]string) string {
	swap := func(i, j int) {
		clock[i], clock[j] = clock[j], clock[i]
	}

	// Algoritmo de ordenação manual (Bubble Sort) pela segunda coluna
	n := len(clock)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if clock[j][0] > clock[j+1][0] {
				swap(j, j+1)
			}
		}
	}

	var r string = "C("
	for i := 0; i < 10; i++ {
		if clock[i][0] != "" {
			var adr string = clock[i][0]
			r += adr[10:] + " -> " + clock[i][1] + ", "
		}
	}
	r = r[:len(r)-2]
	r += ")"
	return r
}

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Please specify at least one address:port!")
		fmt.Println("go run chatBEB.go 127.0.0.1:5001  127.0.0.1:6001   127.0.0.1:7001")
		fmt.Println("go run chatBEB.go 127.0.0.1:6001  127.0.0.1:5001   127.0.0.1:7001")
		fmt.Println("go run chatBEB.go 127.0.0.1:7001  127.0.0.1:6001   127.0.0.1:5001")
		return
	}

	var registro []string
	idp := os.Args[1]
	addresses := os.Args[1:]
	fmt.Println(addresses)

	clockLen := len(os.Args) - 1

	beb := BestEffortBroadcast_Module{
		Req: make(chan BestEffortBroadcast_Req_Message),
		Ind: make(chan BestEffortBroadcast_Ind_Message)}

	//beb.Init(addresses[0])
	beb.InitD(addresses[0], false)

	var vClock [10][2]string

	if clockLen > len(vClock) {
		fmt.Println("allocate more memory to the clock.")
		return
	}

	for i := 0; i < len(addresses); i++ {
		vClock[i][0] = addresses[i]
		vClock[i][1] = "0"
	}

	//fmt.Printf("Clock init: %v\n", formatClock(vClock))

	// enviador de broadcasts
	go func() {

		scanner := bufio.NewScanner(os.Stdin)
		var msg string

		for {
			fmt.Printf("Clock: %v\n", formatClock(vClock))
			//fmt.Println(idp)
			if scanner.Scan() {
				msg = scanner.Text()
				vClock = inc(vClock, idp, 1)
				stringClock := clockToString(vClock)
				msg += "§" + addresses[0] + "§" + stringClock
			}

			req := BestEffortBroadcast_Req_Message{
				Addresses: addresses[0:],
				Message:   msg}
			//fmt.Printf("\nenviado:\n%v\n", vClock)
			beb.Req <- req // ENVIA PARA TODOS PROCESSOS ENDERECADOS NO INICIO
		}
	}()

	// receptor de broadcasts
	go func() {
		for {
			in := <-beb.Ind // RECEBE MENSAGEM DE QUALQUER PROCESSO
			message := strings.Split(in.Message, "§")
			in.From = message[1]
			registro = append(registro, in.Message)
			in.Message = message[0]

			//fmt.Printf("\nrecebido:\n%v\n", message[2])

			var VectorClockMatrix [10][2]string

			VectorClockMatrix = stringToClock(message[2])

			for i := 0; i < clockLen; i++ {
				if VectorClockMatrix[i][1] == "" || VectorClockMatrix[i][1] == " " {
					continue
				}

				for j := 0; j < clockLen; j++ {
					if vClock[j][0] == VectorClockMatrix[i][0] {
						if VectorClockMatrix[i][1] > vClock[j][1] {
							vClock[j][1] = VectorClockMatrix[i][1]
						}
					}
				}

			}

			vClock = inc(vClock, idp, 1)

			// imprime a mensagem recebida na tela
			fmt.Printf("               Message from %v: %v,	  %v\n", in.From, in.Message, formatClock(VectorClockMatrix))
		}
	}()

	blq := make(chan int)
	<-blq
}
