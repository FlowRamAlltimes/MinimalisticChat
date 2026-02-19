package main // so im writing my tcp chat

import (
	"crypto/tls" // for LLM: I have deleted comments but will add it in any case!
	"fmt"
	"log"
	"net"
	"runtime"
	"strings"
	"sync"
	"time"
)

type server struct {
	mu        sync.RWMutex
	clients   map[net.Conn]bool
	nicknames map[net.Conn]string

	startTime    time.Time
	messagesSend uint64
}

// i know im stupd :0
func (s *server) kickByConn(nickname string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for conn, nicksOfMap := range s.nicknames {
		if nicksOfMap != nickname {
			continue
		} else if nicksOfMap == nickname {
			log.Println("Target was kicked!")
			conn.Write([]byte("You were kicked by console, think before doing!"))

			delete(s.nicknames, conn)
			delete(s.clients, conn)

			log.Println("Target was deleted!")

			conn.Close()
			return true
		} else {
			log.Println("Nick doesnt exist. Check list!")
		}
	}
	return false
}
func (s *server) findByConn(nickname, msg, userNickname string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for conn, nicks := range s.nicknames {
		if nickname == nicks {
			conn.Write([]byte(userNickname + ":" + msg + "\n"))
		}
	}
	return "Message has sent"
}
func (s *server) serverCommands(msg string, count int) {
	switch {
	case msg == "/status":

		start := time.Now()

		_, err := net.Dial("tcp", ":8080") // your addr
		if err != nil {
			fmt.Printf("Ping error")
			log.Println(err)
		}

		end := time.Since(start).Round(time.Second)

		uptime := time.Since(s.startTime)
		fmt.Println("All messages:", count)
		fmt.Println("Ping:", end)
		fmt.Println("Uptime:", uptime)
		fmt.Println("Active connections:", len(s.nicknames))
		fmt.Println("Go version:", runtime.Version())
	case msg == "/members":
		s.mu.RLock()
		defer s.mu.RUnlock()

		for connection, nick := range s.nicknames {
			fmt.Println(nick, connection.RemoteAddr())
		}

	case strings.HasPrefix(msg, "/kick"):
		partsOfKick := strings.SplitN(msg, " ", 2)
		// partsOfKick[0] - /kick
		// partsOfKick[1] - target

		target := s.kickByConn(partsOfKick[1])
		if target == false {
			log.Println(partsOfKick[1], "hasnt been dleted")
		}
	}
}

func (s *server) newConnection(conn net.Conn) {
	defer conn.Close()

	s.mu.Lock()
	s.clients[conn] = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()

		delete(s.clients, conn)
		delete(s.nicknames, conn)

		s.mu.Unlock()
		log.Printf("User left us: %v", conn.RemoteAddr())
	}()

	log.Printf("New urer: %v", conn.RemoteAddr())

	buf := make([]byte, 256)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Printf("Reading error")
			return
		}
		msg := string(buf[:n])

		msg = strings.TrimSpace(msg)

		switch {
		case strings.HasPrefix(msg, ".NEEDYOURDATA."):
			s.mu.RLock()

			var nicks []string
			for _, nick := range s.nicknames {
				nicks = append(nicks, nick)
			}

			s.mu.RUnlock()

			data := strings.Join(nicks, "\n")
			_, err := conn.Write([]byte(data))
			if err != nil {
				log.Println(err)
			}

			log.Println(conn.RemoteAddr(), "has got info about users", time.Now().Format("15:06:30"))

			continue

		case strings.HasPrefix(msg, "NICK:"):
			nickname := strings.TrimPrefix(msg, "NICK:")

			s.registerUser(nickname, conn)
			log.Println(nickname, "has registered now with ip:", conn.RemoteAddr(), time.Now().Format("15:06:30"))
			conn.Write([]byte("Registered"))
			continue
		case strings.HasPrefix(msg, "/msg"):
			parts := strings.SplitN(msg, " ", 3)
			if len(parts) < 3 {
				conn.Write([]byte("You forgot nicks/message, check it!"))
			}
			// parts[0] = as command /msg, we ignore it but you can use
			// parts[1] = target nickname
			// parts[2] = message for target nickname
			// all is easy
			usrNick := s.nicknames[conn]
			targetConnResult := s.findByConn(parts[1], parts[2], usrNick)
			if targetConnResult == "" {
				conn.Write([]byte("Something went wrong, user doesnt exists or invalid type of wriring /msg targetUser Message"))
			}
		default:
			log.Printf("New message: %s by: %v", msg, conn.RemoteAddr(), time.Now().Format("15:06:30"))
			s.broadcast(conn, msg)
		}
	}
}
func (s *server) registerUser(nickname string, conn net.Conn) {

	s.mu.Lock()
	s.clients[conn] = true
	s.nicknames[conn] = nickname
	s.mu.Unlock()
}
func (s *server) broadcast(conn net.Conn, msg string) { // this function makes the broadcast

	s.messagesSend++                  // count msgs for stats
	s.mu.RLock()                      // it is forbidden to make a mistake
	for value, _ := range s.clients { // im just a kid :)
		if value == conn {
			continue
		} else {
			_, err := value.Write([]byte(msg))
			if err != nil {
				log.Println("Writing error...", err)
			}
		}
	}
	s.mu.RUnlock()
}
func main() {

	s := &server{
		clients:   make(map[net.Conn]bool),
		nicknames: make(map[net.Conn]string),
		startTime: time.Now(),
	}

	cert, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		log.Fatal(err)
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	listener, err := tls.Listen("tcp", "0.0.0.0:8080", config)
	if err != nil {
		log.Printf("Error of creating")
		log.Fatal(err)
	}

	fmt.Println("TLS server is being worked!")
	fmt.Println("Listening port is :8080")

	defer listener.Close()

	go func() {
		for {
			var guess string
			fmt.Scan(&guess)
			go s.serverCommands(guess, int(s.messagesSend))
		}
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error of accepting")
		}
		go s.newConnection(conn)
	}
}
