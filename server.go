// Documentation:
// This project shows at least the worst part of package net.
// So the purpose of this code is to understand how tcp and ip works on high-level programming
// Full explanation is showed in DOCS.md in my GitHub repository
// Thanks for reading
// Process of working:
// Server is running on port that was parsed with -p, -pRst(for http server)
// After client connects to it with registration, so then you can start chatting with your friends or smb else.
package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"flag"
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

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const logfile = "server.log"    // files !WARNING! CHECK IF YOUR FILES USELESS BEFORE OR IT WILL BE PERFECTLY THAT TREY ARE NOT EXIST
const banlist = "banlist.txt"   // IF YOU HAVE THE SAME ONES IT WILL BETTER IF YOU RENAME THEM
const mutelist = "mutelist.txt" // ITS FOR MUTES

type server struct {
	// part of infrastructure
	mu            sync.RWMutex
	clients       map[net.Conn]bool
	banaddresses  map[string]string
	muteaddresses map[string]string
	nicknames     map[net.Conn]string // find user by ip in O(1)
	addresses     map[string]net.Conn // find connection address by nick in O(1)
	// part of metrics
	StartTime    time.Time // uptime check
	MessagesSend uint64    // all messages counting
	// errors counting
	ConnectionErrors  uint64
	WritingErrors     uint64
	CertificateErrors uint64
	ReadingErrors     uint64
	AcceptErrors      uint64
}

var (
	messagesSent = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "messages_sent",
		Help: "Total number of messages sent.",
	}, []string{"messages_all"})

	connErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "conn_err",
		Help: "Total connectinon errors.",
	}, []string{"conn_err"})

	writingErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "writing_err",
		Help: "Total writing errors.",
	}, []string{"writing_err"})

	certificateErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "cert_err",
		Help: "Total certificate errors.",
	}, []string{"cert_err"})

	readingErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "reading_err",
		Help: "Total reading errors.",
	}, []string{"reading_err"})

	acceptErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "accept_err",
		Help: "Total accept errors.",
	}, []string{"accept_err"})

	errors_total = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "errors_total",
		Help: "Total errors.",
	}, []string{"errors_total"})
)

// copypasted functions btw
func mainpage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<a href='/metrics'>prometheus metrics</a>")
	fmt.Fprintf(w, "<a href='/status'>server status</a>")
	fmt.Fprintf(w, "<a href='/status/server'>server information(in process of making)</a>")
}
func (s *server) unMuteByIp(targetIP string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := os.Remove(mutelist)
	if err != nil {
		log.Printf("Error when removing/deleting file/user")
		return false
	}
	f, err := os.OpenFile(mutelist, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("Error of appending/creating file")
		return false
	}
	defer f.Close()
	for val, _ := range s.muteaddresses {
		if val == targetIP {
			delete(s.muteaddresses, targetIP)
		} else {
			continue // ignoring others(it is possible to do update)
		}
	}
	for value, _ := range s.muteaddresses {
		io.WriteString(f, value) // sorry for bad naming
	}
	return true
}
func (s *server) unBanByIp(targetIP string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := os.Remove(banlist)
	if err != nil {
		log.Printf("Error when removing/deleting file/user")
		return false
	}
	f, err := os.OpenFile(banlist, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("Error of appending/creating file")
		return false
	}
	defer f.Close()
	for val, _ := range s.banaddresses {
		if val == targetIP {
			delete(s.banaddresses, targetIP)
		} else {
			continue // ignoring others(it is possible to do update)
		}
	}
	for value, _ := range s.banaddresses {
		io.WriteString(f, value) // sorry for bad naming
	}
	return true
}
func (s *server) uploadMapForMute(conn net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	f, err := os.Open(mutelist)
	if err != nil {
		log.Printf("Error while opening file")
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		ip := strings.TrimSpace(scanner.Text())
		if ip != "" {
			s.muteaddresses[ip] = "mutedByConsole"
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading file: %v", err)
	}
}
func (s *server) muteByName(nick string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	addr := s.addresses[nick]
	addrWithout, _, err := net.SplitHostPort(addr.RemoteAddr().String())
	if err != nil {
		log.Printf("Error %v", err)
	}
	f, error := os.OpenFile(mutelist, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if error != nil {
		log.Printf("Error opening file %v", error)
	}
	defer f.Close()

	io.WriteString(f, addrWithout)
	addr.Close()
	return true
}

func (s *server) checkIpByBanning(conn net.Conn) bool {
	s.mu.RLock()                                                         // locks
	defer s.mu.RUnlock()                                                 // unlocks when func is done
	remoteAddress, _, _ := net.SplitHostPort(conn.RemoteAddr().String()) // remote address
	// checking for entrance
	_, isBanned := s.banaddresses[remoteAddress]

	if isBanned {
		log.Printf("IP %s is in ban-list. Closing connection.", remoteAddress)
		fmt.Fprintf(conn, "!_YOU HAD BEEN BANNED ON OUR SERVER_!")
		conn.Close()
		return true
	}

	return false
}
func (s *server) uploadMapForBan(conn net.Conn) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	remote, _, _ := net.SplitHostPort(conn.RemoteAddr().String())

	file, err := os.Open(banlist) // opening file
	if err != nil {
		log.Printf("Error of opening file")
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		ip := strings.TrimSpace(scanner.Text())
		if ip != "" {
			// ban persone
			s.banaddresses[ip] = "bannedByConsole"
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading file: %v", err)
	}

	return remote
}

func (s *server) banWithIp(ip net.Conn) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	stringIp, _, _ := net.SplitHostPort(ip.RemoteAddr().String())

	f, err := os.OpenFile(banlist, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("Error of opening banfile")
	}
	defer f.Close()

	io.WriteString(f, stringIp)
	ip.Write([]byte("You've been banned by ip address, see ya!"))

	delete(s.clients, ip)
	delete(s.nicknames, ip)
	for nick, c := range s.addresses {
		if c == ip {
			delete(s.addresses, nick)
		}
	}

	ip.Close()
	return true
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
	fmt.Fprintf(w, "Hello, its info endpoint of my chat, just test btw. ")
	fmt.Fprintf(w, "<p>I think that it will be helpful in future.</p>")
	fmt.Fprintf(w, "<p>So its my online now: %v</p>", len(s.nicknames))
	fmt.Fprintf(w, "<p><a href='https://github.com/FlowRamAlltimes/MinimalisticChat'>It's my GitHub repository</a></p>")
}
func (s *server) serverStatusJson(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	stats := struct {
		Uptime                string `json:"uptime"`
		AllMessages           uint64 `json:"all_messages"`
		Errors                uint64 `json:"errors_since_start"`
		Connect_errors        uint64 `json:"connection_errors"`
		Writing_errors        uint64 `json:"writing_errors"`
		Certificate_errors    uint64 `json:"certificate_errors"`
		Reading_or_EOF_errors uint64 `json:"reading_errors"`
		Accepting_errors      uint64 `json:"accepting_errors"`
		AllUsersInActive      int    `json:"active_users"`
	}{
		Uptime:                time.Since(s.StartTime).Round(time.Second).String(),
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
				delete(s.addresses, nick)
				delete(s.nicknames, conn)
				delete(s.clients, conn)
				// appending new
				s.clients[connect] = true
				s.nicknames[connect] = nick
				s.addresses[nick] = connect
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
			delete(s.addresses, nicksOfMap)

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
func (s *server) serverCommands(msg string, count, connectErrors, readingErrors, acceptingError, certError, writingError int, addr string) {
	switch {
	case msg == "/status":

		start := time.Now()

		_, err := net.Dial("tcp", addr)
		if err != nil {
			fmt.Printf("Ping error")
			errors_total.WithLabelValues("conn_err").Inc()
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
		fmt.Println("Achitecture:", runtime.GOARCH)
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
			log.Println(partsOfKick[1], "hasn't been deleted")
		}
	case strings.HasPrefix(msg, "/ban"):
		partsOfBan := strings.SplitN(msg, " ", 2)
		if len(partsOfBan) < 2 {
			log.Printf("Follow the rules of /ban ActiveNickname")
		}
		conn := s.addresses[partsOfBan[1]]
		res := s.banWithIp(conn)
		if res == true {
			log.Printf("Enemy has been banned by console: %s", partsOfBan[1])
		} else {
			log.Printf("Error while banning %s", partsOfBan[1])
		}
	case strings.HasPrefix(msg, "/mute"):
		muteParts := strings.SplitN(msg, " ", 2)
		// muteParts[0] - /msg
		// muteParts[1] - nickName
		result := s.muteByName(muteParts[1])
		if result == true {
			fmt.Printf("Target was mute dby console %v", time.Now().Local())
		}
	case strings.HasPrefix(msg, "/unmute"):
		unmutePrts := strings.SplitN(msg, " ", 2)
		// unmutePrts[0] - /unmute(ignoring)
		// unmutePrts[1] - targetIP
		resultOfUnmute := s.unMuteByIp(unmutePrts[1])
		if resultOfUnmute == true {
			log.Printf("Unmuting has been made sucessfully!")
		} else {
			log.Printf("An error has been here")
		}
	case strings.HasPrefix(msg, "/unban"):
		unbanPrts := strings.SplitN(msg, " ", 2)
		// unbanPrts[0] - /unmute(ignoring)
		// unbanPrts[1] - targetIP
		resultOfUnban := s.unBanByIp(unbanPrts[1])
		if resultOfUnban == true {
			log.Printf("Unbanning has been made sucessfully!")
		} else {
			log.Printf("An error has been here")
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
		for nick, c := range s.addresses {
			if c == conn {
				delete(s.addresses, nick)
			}
		}

		s.mu.Unlock()
		log.Printf("User left us: %v", conn.RemoteAddr())
	}()

	log.Printf("New urer: %v", conn.RemoteAddr())

	buf := make([]byte, 256)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Printf("Reading error")
			errors_total.WithLabelValues("reading_err").Inc()
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
				errors_total.WithLabelValues("writing_err").Inc()
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
			// all is easy#

			usrNick := s.nicknames[conn]
			targetConnResult := s.findByConn(parts[1], parts[2], usrNick)
			if targetConnResult == "" {
				conn.Write([]byte("Something went wrong, user doesnt exists or invalid type of wriring /msg targetUser Message"))
			}
		case strings.HasPrefix(msg, ".1wannachangen1ck."):
			newNickParts := strings.SplitN(msg, " ", 2)

			// newNickParts[1] = new nickname

			if len(newNickParts) < 2 {
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
	// registring user
	s.mu.Lock()
	s.clients[conn] = true
	s.nicknames[conn] = nickname
	s.addresses[nickname] = conn
	s.mu.Unlock()
}
func (s *server) broadcast(conn net.Conn, msg string) { // this function makes the broadcast
	flag := false
	s.mu.RLock()
	defer s.mu.RUnlock()
	addrWithoutPort, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
	// brilliant idea with flag
	for connectForbidenMute, _ := range s.muteaddresses {
		if addrWithoutPort == connectForbidenMute {
			flag = true
		}
	}

	messagesSent.WithLabelValues("messages_all").Inc() // metrics
	s.MessagesSend++                                   // other info by /status

	for value, _ := range s.clients {
		if flag == true {
			conn.Write([]byte("YOU ARE MUTED, STUPID"))
			continue
		}
		if value == conn {
			continue
		} else {
			_, err := value.Write([]byte(msg))
			if err != nil {
				errors_total.WithLabelValues("writing_err").Inc()
				s.WritingErrors++
				log.Println("Writing error...", err)
			}
		}
	}
}
func main() {
	address := flag.String("addr", "", "use for set address of server")
	portForChat := flag.Int("p", 0, "use for setting port for chat")
	portForRESTapi := flag.Int("pRst", 0, "use for setting port for REST API")

	flag.Parse()

	switch {
	case *address == "":
		fmt.Printf("U must set address for chat with -addr")
		return
	case *portForChat == 0:
		fmt.Printf("U must set port for chat with -p")
		return
	case *portForRESTapi == 0:
		fmt.Printf("U must set port for REST API with -pRst")
		return
	}

	portForChatStr := strconv.Itoa(*portForChat)
	addr := fmt.Sprintf(*address + ":" + portForChatStr)

	portForRESTapiStr := strconv.Itoa(*portForRESTapi)
	addrForRest := fmt.Sprintf(*address + ":" + portForRESTapiStr)

	f, err := os.OpenFile(logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Error of opening log-file: %v", logfile, err)
	}
	defer f.Close()

	mw := io.MultiWriter(os.Stdout, f)

	log.SetOutput(mw)

	log.Println("Log system has started")

	s := &server{
		clients:       make(map[net.Conn]bool),
		nicknames:     make(map[net.Conn]string),
		addresses:     make(map[string]net.Conn),
		banaddresses:  make(map[string]string),
		muteaddresses: make(map[string]string),
		StartTime:     time.Now(),
	}

	// http server stats in json

	mux := http.NewServeMux() // init new router, soon it will customed

	go func() { // routes
		mux.HandleFunc("/", mainpage)
		mux.Handle("/metrics", promhttp.Handler())                          // metrics endpoint for prometheus
		mux.HandleFunc("/status", s.serverStatusJson)                       // first endpoint
		mux.HandleFunc("/api", s.infoPageOfChat)                            // and the second
		mux.HandleFunc("/status/server", s.serverinfo)                      // third
		log.Printf("Server monitor is aviable on: https://%v", addrForRest) // yeah I finally understood how to run server with https protocol

		portForRESTapiInListenTLS := fmt.Sprintf("0.0.0.0:" + portForRESTapiStr)

		APIserver := &http.Server{ // custom running
			Addr:         portForRESTapiInListenTLS,
			Handler:      mux,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
		}

		APIserver.ListenAndServeTLS("server.crt", "server.key") // thats convenient as I can use I self-created cert for here
	}()
	cert, err := tls.LoadX509KeyPair("server.crt", "server.key") // and here, btw i used OpenSSL certs
	if err != nil {
		certificateErrors.WithLabelValues("cert_err").Inc()
		s.CertificateErrors++
		log.Fatal(err)
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	portForChatInListener := fmt.Sprintf("0.0.0.0:" + *&portForChatStr)

	listener, err := tls.Listen("tcp", portForChatInListener, config)
	if err != nil {
		log.Printf("Error of creating")
		log.Fatal(err)
	}

	fmt.Println("TLS server is being worked!")
	fmt.Printf("Listening port is: %v", portForChatStr)

	defer listener.Close()

	go func() {
		for {
			var guess string
			commandReader := bufio.NewReader(os.Stdin)
			guess, _ = commandReader.ReadString('\n')
			guess = strings.TrimSpace(guess)
			go s.serverCommands(guess, int(s.MessagesSend), int(s.ConnectionErrors), int(s.ReadingErrors), int(s.AcceptErrors), int(s.CertificateErrors), int(s.WritingErrors), addr)
		}
	}()

	for {
		conn, err := listener.Accept()

		if err != nil {
			acceptErrors.WithLabelValues("accept_err").Inc()
			s.AcceptErrors++
			log.Printf("Error of accepting")
		}
		s.uploadMapForMute(conn)
		s.uploadMapForBan(conn)
		if s.checkIpByBanning(conn) == false {
			go s.newConnection(conn)
		}
	}
}

// I dont know how big it will be soon
// cool, so i think that i'll grow up soon but who can know the truth if it isnt me . . .
