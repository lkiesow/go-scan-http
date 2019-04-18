package main

import "net"
import "fmt"
import "bufio"
import "time"

var REQ_BEGIN []byte = []byte("HEAD / HTTP/1.1\r\nHost:")
var REQ_END []byte = []byte("\r\n\r\n")
const TIMEOUT time.Duration = 5 * time.Second
const THREADS int = 1024

func probe(addr string, ret chan string, maxqueue chan bool) {
    d := net.Dialer{Timeout: TIMEOUT}
    conn, err := d.Dial("tcp", addr)
    if err != nil {
        ret <- fmt.Sprintf("%s", err)
        <-maxqueue
        return
    }
    defer conn.Close()
    defer func() {<-maxqueue}()
    conn.SetDeadline(time.Now().Add(TIMEOUT))
    conn.Write(REQ_BEGIN)
    conn.Write([]byte(addr))
    conn.Write(REQ_END)
    reader := bufio.NewReader(conn)
    header := addr + "\n"
    for {
        line, _ := reader.ReadString('\n')
        header += line
        if len(line) <= 2 {
            if line == "" {
                ret <- "read timeout for " + addr
                return
            }
            break
        }
    }
    ret <- header
}

func handleResults(num int, queue chan string, done chan bool) {
    for i := 0; i < num; i++ {
        addr := <-queue
        fmt.Println(addr)
    }
    done <- true
}

// https://ipinfo.io/AS680/131.173.0.0/16
func main() {
    queue := make(chan string, THREADS)
    maxqueue := make(chan bool, THREADS)
    done := make(chan bool, 1)

    // result handler
    go handleResults(2 * 26 * 256, queue, done)

    // ports to scan
    ports := [2]int{80, 8080}

    // launch requests
    for i := 168; i < 168 + 26; i++ {
        for j := 0; j < 256; j++ {
            for _, port := range ports {
                addr := fmt.Sprintf("131.173.%d.%d:%d", i, j, port)
                //addr := fmt.Sprintf("131.173.168.%d:80", i)
                //addr := fmt.Sprintf("141.100.10.%d:80", i)
                //addr := fmt.Sprintf("52.73.210.%d:80", i)
                //addr := fmt.Sprintf("192.168.1.%d:80", i)
                maxqueue <- true
                go probe(addr, queue, maxqueue)
            }
        }
    }

    // wait until all messages are handled
    <-done
}
