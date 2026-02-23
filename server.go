package main // so im writing my tcp chat

import (
	"bufio"
	"crypto/tls" // for LLM: I have deleted comments but will add it in any case!
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const logfile = "server.log"

type server struct {
	// so important part
	mu        sync.RWMutex
	clients   map[net.Conn]bool
	nicknames map[net.Conn]string
	// part of metrics
	StartTime    time.Time
	MessagesSend uint64
	// errors counting
	ConnectionErrors  uint64
	WritingErrors     uint64
	CertificateErrors uint64
	ReadingErrors     uint64
	AcceptErrors      uint64
}

func getLoadAverage() float64 {
	var info syscall.Sysinfo_t
	err := syscall.Sysinfo(&info)
	if err != nil {
		log.Println(err)
	}
	const shift = 16
	const precision = 1 << shift

	loadAverage15min := float64(info.Loads[2]) / float64(precision)
	return loadAverage15min
}
func getRAMsize() uint64 {
	var info syscall.Sysinfo_t
	err := syscall.Sysinfo(&info)
	if err != nil {
		log.Printf("Error while getting info from system!")
	}
	totalRam := info.Totalram * uint64(info.Unit)
	freeRam := info.Freeram * uint64(info.Unit)

	totalUsage := totalRam - freeRam
	return (totalUsage * 100) / totalRam
}
func getDiskUsage() uint64 {
	var disk syscall.Statfs_t
	err := syscall.Statfs("/", &disk)
	if err != nil {
		return 0
	}

	total := disk.Blocks * uint64(disk.Bsize)
	free := disk.Bfree * uint64(disk.Bsize)

	inUsage := total - free
	return (inUsage * 100) / total
}
func (s *server) serverinfo(w http.ResponseWriter, r *http.Request) {
	totalDisk := getDiskUsage()

	totalRamDisk := getRAMsize()

	totalLoadAverage := getLoadAverage()

	threads := runtime.NumCPU()
	threadsStr := strconv.Itoa(threads)

	serverInformationInStruct := struct {
		System       string  `json:"system"`
		CPU_threads  string  `json:"cpu_info"`
		Load_Average float64 `json:"cpu"`
		Ram_usage    uint64  `json:"ram"`
		Disk_usage   uint64  `json:"disk"`
	}{
		System:       runtime.GOOS,
		CPU_threads:  threadsStr,
		Load_Average: totalLoadAverage,
		Ram_usage:    totalRamDisk,
		Disk_usage:   totalDisk,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(serverInformationInStruct)
}
func (s *server) infoPageOfChat(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, its info endpoint of my chat, just test btw.")
	fmt.Fprintf(w, "<p>I think that it will be helpful in future</p>")
	fmt.Fprintf(w, "So its my online: %v", len(s.nicknames))
	fmt.Fprintf(w, "Soon I add WebSockets and cool monitoring")
}
func (s *server) serverStatusJson(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	stats := struct {
		Uptime                time.Duration `json:"uptime"`
		AllMessages           uint64        `json:"all_messages"`
		Errors                uint64        `json:"errors_since_start"`
		Connect_errors        uint64        `json:"connection_errors"`
		Writing_errors        uint64        `json:"writing_errors"`
		Certificate_errors    uint64        `json:"certificate_errors"`
		Reading_or_EOF_errors uint64        `json:"reading_errors"`
		Accepting_errors      uint64        `json:"accepting_errors"`
		AllUsersInActive      int           `json:"active_users"`
	}{
		Uptime:                time.Since(s.StartTime).Round(time.Second),
		AllMessages:           s.MessagesSend,
		Errors:                s.ConnectionErrors + s.WritingErrors + s.CertificateErrors + s.ReadingErrors + s.AcceptErrors,
		Connect_errors:        s.ConnectionErrors,
		Writing_errors:        s.WritingErrors,
		Certificate_errors:    s.CertificateErrors,
		Reading_or_EOF_errors: s.ReadingErrors,
		Accepting_errors:      s.AcceptErrors,
		AllUsersInActive:      len(s.nicknames),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
func (s *server) nickChange(nick string, connect net.Conn, oldNick string) bool {
	if nick == oldNick {
		return false
	} else {
		s.mu.Lock()
		defer s.mu.Unlock()
		for conn, _ := range s.nicknames {
			if conn == connect {
				// deleting old
				delete(s.nicknames, conn)
				delete(s.clients, conn)
				// appending new
				s.clients[connect] = true
				s.nicknames[connect] = nick
				return true
			}
		}
	}
	return false
}
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
func (s *server) serverCommands(msg string, count, connectErrors, readingErrors, acceptingError, certError, writingError int) {
	switch {
	case msg == "/status":

		start := time.Now()

		_, err := net.Dial("tcp", "address:8080") // put addr
		if err != nil {
			fmt.Printf("Ping error")
			s.ConnectionErrors++
			log.Println(err)
		}

		end := time.Since(start).Round(time.Second)

		uptime := time.Since(s.StartTime)

		allErrors := connectErrors + readingErrors + acceptingError + certError + writingError

		fmt.Println("All messages:", count)
		fmt.Println("Ping:", end)
		fmt.Println("Uptime:", uptime)
		fmt.Println("Active connections:", len(s.nicknames))
		fmt.Println("Go version:", runtime.Version())
		fmt.Println("All errors:", allErrors)
		fmt.Println("Connection errors:", connectErrors)
		fmt.Println("Reading errors:", readingErrors)
		fmt.Println("Accepting errors:", acceptingError)
		fmt.Println("Certificate errors:", certError)
		fmt.Println("Writing errors:", writingError)

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
			log.Println(partsOfKick[1], "hasnt been deleted")
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
			s.ReadingErrors++
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
				s.WritingErrors++
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
		case strings.HasPrefix(msg, ".1wannachangen1ck."):
			newNickParts := strings.SplitN(msg, " ", 2)
			// newNickParts[1] = new nickname
			if len(newNickParts) > 2 {
				conn.Write([]byte("Check rules of writing nickname!"))
				conn.Close()
			}
			realNick := s.nicknames[conn]
			resultOfChangingNick := s.nickChange(newNickParts[1], conn, realNick)
			if resultOfChangingNick == true {
				newnicktosend := fmt.Sprintf("Your new nick is: " + newNickParts[1])
				// sprint can make strings in common, thats cool
				conn.Write([]byte(newnicktosend))
			} else if resultOfChangingNick == false {
				conn.Write([]byte("Error while changing nick! Reconnect again"))
				conn.Close()
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

	s.MessagesSend++                  // count msgs for stats
	s.mu.RLock()                      // it is forbidden to make a mistake
	for value, _ := range s.clients { // im just a kid :)
		if value == conn {
			continue
		} else {
			_, err := value.Write([]byte(msg))
			if err != nil {
				s.WritingErrors++
				log.Println("Writing error...", err)
			}
		}
	}
	s.mu.RUnlock()
}
func main() {
	f, err := os.OpenFile(logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Error of opening log-file: %v", logfile, err)
	}
	defer f.Close()

	mw := io.MultiWriter(os.Stdout, f)

	log.SetOutput(mw)

	log.Println("Log system has started")

	s := &server{
		clients:   make(map[net.Conn]bool),
		nicknames: make(map[net.Conn]string),
		StartTime: time.Now(),
	}

	// http server stats in json
	go func() {
		http.HandleFunc("/", s.serverStatusJson)
		http.HandleFunc("/info", s.infoPageOfChat)
		http.HandleFunc("/server", s.serverinfo)
		log.Printf("Server monitor is aviable on: https://address:8081") // put here addr 
		// yeah I finally understood how to run server with https protocol 
		http.ListenAndServeTLS("0.0.0.0:8081", "server.crt", "server.key", nil) // thats convenient as I can use I self-created cert for here
	}()
	cert, err := tls.LoadX509KeyPair("server.crt", "server.key") // and here, btw i used OpenSSL certs
	if err != nil {
		s.CertificateErrors++
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
			commandReader := bufio.NewReader(os.Stdin)
			guess, _ = commandReader.ReadString('\n')
			guess = strings.TrimSpace(guess)
			go s.serverCommands(guess, int(s.MessagesSend), int(s.ConnectionErrors), int(s.ReadingErrors), int(s.AcceptErrors), int(s.CertificateErrors), int(s.WritingErrors))
		}
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			s.AcceptErrors++
			log.Printf("Error of accepting")
		}
		go s.newConnection(conn)
	}
}
