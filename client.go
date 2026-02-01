package main                // Hello

import (
        "bufio"        //for writing 
        "fmt"          //i think you know
        "log"          //logging
        "net"          //cool place for tcp/udp/ip
        "os"           //os for checks env
        "strings"      //convertation into strings
        "time"         //ping and other things
)

func readServerMessages(conn net.Conn) {
        buf := make([]byte, 1024)
        for {
                n, err := conn.Read(buf)
                if err != nil {
                        log.Printf("EOF danger, reconnenct again")
                        return
                }
                msg := string(buf[:n])
                fmt.Printf(">> %s\nВы: ", msg)
        }
}
func healthCheck() {
        start := time.Now()
        _, err := net.DialTimeout("tcp", "localhost:8080", 1*time.Second)
        if err != nil {
                log.Printf("connection error")
                return
        }
        result := time.Since(start)
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
                }
                _, err := conn.Write([]byte(text + "\n"))
                if err != nil {
                        log.Printf("Sending error, try contact @ramhely")
                }
        }
}
