package main // Hello

import (
	"bufio"   //for writing
	"fmt"     //i think you know
	"log"     //logging
	"net"     //cool place for tcp/udp/ip
	"os"      //os for checks env
	"strings" //convertation into strings
	"time"    //ping and other things
)

func info() {
	fmt.Println("v1.1207")
	fmt.Println("WELCOME TO MY TCP CHAT")
	fmt.Println("It's wonderful place where you can talk with your friends")
	fmt.Println("If you are fan of old typed chats, I can show you it")
	fmt.Println("Send your first message!")
	fmt.Println("Use /help for list of commands")
}
func myIp(conn net.Conn) {
	fmt.Println("Your IP address is:", conn.RemoteAddr())
}
func exitFunc() {
	os.Exit(0)
}
func helplist() { // help function its very cool
	fmt.Println("=============================")
	fmt.Println("Hello you've used /help")
	fmt.Println("/health - check server's state")
	fmt.Println("/list - check all active members")
	fmt.Println("/help - get help about server")
	fmt.Println("/quit or /exit - leave chat")
	fmt.Println("=============================")
}
func readServerMessages(conn net.Conn) {
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Printf("EOF danger, reconnenct again")
			return
		}
		msg := string(buf[:n])
		fmt.Printf(">> %s\n", msg)
	}
}
func healthCheck() {
	start := time.Now()
	_, err := net.DialTimeout("tcp", "localhost:8080", 1*time.Second)
	if err != nil {
		log.Printf("connection error")
		return
	}
	result := time.Since(start).Round(time.Millisecond)
	fmt.Println("Your ping is:", result)
	fmt.Println("Server is OK!")
}

func list(conn net.Conn) {
	_, err := conn.Write([]byte(".NEEDYOURDATA.")) // .NEEDYOURDATA.
	if err != nil {
		log.Printf("Nothing happened")
		return
	}
	newbuf := make([]byte, 256)
	n, err := conn.Read(newbuf)
	if err != nil {
		log.Printf("Reading error")
		return
	}
	msg := string(newbuf[:n])
	fmt.Println(msg)
}

func main() {
	conn, err := net.DialTimeout("tcp", "localhost:8080", 2*time.Second)
	if err != nil {
		log.Printf("Connection error, try again later")
		log.Printf("If it does not help, contact me in tg: @ramhely")
		return
	}
	defer conn.Close()

	fmt.Println("Please enter with nickname:")
	nickreader := bufio.NewReader(os.Stdin)
	msg, _ := nickreader.ReadString('\n')
	msg = strings.TrimSpace(msg)
	nick := msg

	conn.Write([]byte("NICK:" + msg + "\n"))

	go readServerMessages(conn)

	for {
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		if text == "" {
			break
		} else if text == "/health" {
			go func() {
				healthCheck()
			}()
		} else if text == "/list" {
			go func() {
				list(conn)
			}()
		} else if text == "/help" {
			go func() {
				helplist()
			}()
		} else if text == "/quit" || text == "/exit" {
			go func() {
				exitFunc()
			}()
		} else if text == "/ip" {
			go func() {
				myIp(conn)
			}()
		} else if text == "/info" {
			go func() {
				info()
			}()
		}
		_, err := conn.Write([]byte(nick + ":" + text + "\n"))
		if err != nil {
			log.Printf("Sending error, try contact @ramhely")
		}
	}
}
