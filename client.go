package main // Hello

import (
	"bufio" //for writing
	"crypto/tls"
	"crypto/x509"
	_ "embed"
	"fmt"     //i think you know
	"log"     //logging
	"net"     //cool place for tcp/udp/ip
	"os"      //os for checks env
	"strings" //convertation into strings
	"time"    //ping and other things
)

func info() {
	fmt.Println("v1.5")
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
	buf := make([]byte, 256)
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
	_, err := net.DialTimeout("tcp", ":8080", 1*time.Second) // addr
	if err != nil {
		log.Printf("connection error")
		fmt.Println(err)
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

//go:embed ca.crt
var embeddedCert []byte

func main() {
	caCert := embeddedCert

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		log.Fatal("We couldnt add crt: file is invalid or isnt PEM")
	}

	if len(caCert) == 0 {
		panic("Cert is unaviable or isnt in workdir!")
	}
	config := &tls.Config{
		RootCAs:            caCertPool,
		ServerName:         "", // addr
		InsecureSkipVerify: false,
	}
	conn, err := tls.Dial("tcp", ":8080", config) // addr
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
			go healthCheck()
			continue
		} else if text == "/list" {
			go func() {
				list(conn)
			}()
			continue
		} else if text == "/help" {
			go func() {
				helplist()
			}()
			continue
		} else if text == "/quit" || text == "/exit" {
			go func() {
				exitFunc()
			}()
			continue
		} else if text == "/ip" {
			go func() {
				myIp(conn)
			}()
			continue
		} else if text == "/info" {
			go func() {
				info()
			}()
			continue
		} else if strings.HasPrefix(text, "/msg") {
			_, err := conn.Write([]byte(text))
			if err != nil {
				log.Printf("Error while sending private msg!")
				continue
			}
			continue
		}
		_, err := conn.Write([]byte(nick + ":" + text + "\n"))
		if err != nil {
			log.Printf("Sending error, try contact @ramhely")
		}
	}
}
