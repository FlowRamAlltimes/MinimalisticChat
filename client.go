package main // Hello

import (
	"bufio" //for writing
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt" //i think you know
	"log" //logging
	"math/rand/v2"
	"net" //cool place for tcp/udp/ip
	"os"  //os for checks env
	"strconv"
	"strings" //convertation into strings
	"time"    //ping and other things
)

func dicefunc() {
	randNum := 0
	for {
		randNum = rand.IntN(6) + 1
		if randNum != 0 {
			break
		}
	}
	randNumStr := strconv.Itoa(randNum)
	randNumStrUpd := fmt.Sprintf("Your number is: %v", randNumStr)
	fmt.Println(randNumStrUpd)
}

func changeNameWhileOnline(newname string, conn net.Conn) {
	conn.Write([]byte(".1wannachangen1ck." + newname))
	fmt.Printf("Changing nick...")
}
func info() {
	fmt.Println("v1.7.1")
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
func healthCheck(addr, port string) {
	fullAddress := fmt.Sprintf(addr + ":" + port)

	start := time.Now()
	_, err := net.DialTimeout("tcp", fullAddress, 1*time.Second)
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

func main() {
	addr := flag.String("addr", "", "use -addr for connecting")
	port := flag.String("p", "", "use -p for connecting")

	flag.Parse()

	fullAddress := fmt.Sprintf(*addr + ":" + *port)

	caCert, err := os.ReadFile("ca.crt")
	if err != nil {
		log.Printf("error while reading cert %v", err)
		return
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		log.Fatal("We couldnt add crt: file is invalid or isnt PEM")
	}

	if len(caCert) == 0 {
		panic("Cert is unaviable or isnt in workdir!")
	}
	config := &tls.Config{
		RootCAs:            caCertPool,
		ServerName:         *addr,
		InsecureSkipVerify: false,
	}
	conn, err := tls.Dial("tcp", fullAddress, config)
	if err != nil {
		log.Printf("Connection error, try again later, %s", err)
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
			go healthCheck(*addr, *port)
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
		} else if strings.HasPrefix(text, "/change") {
			go func() {
				newNickNameForChat := strings.SplitN(text, " ", 2)
				if len(newNickNameForChat) < 2 {
					log.Printf("Use /change normally")
				}
				changeNameWhileOnline(newNickNameForChat[1], conn)
			}()
			continue
		} else if text == "/dice" {
			go dicefunc()
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
			log.Printf("Sending error %s", err)
		}
	}
}
