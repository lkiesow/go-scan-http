package main

import "net"
import "fmt"
import "bufio"
import "time"

var REQ_BEGIN []byte = []byte("HEAD / HTTP/1.1\r\nHost:")
var REQ_END []byte = []byte("\r\n\r\n")
const TIMEOUT time.Duration = 5 * time.Second
const THREADS int = 512

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
    scan := parseArgs()

    fmt.Println(scan)

    queue := make(chan string, THREADS)
    maxqueue := make(chan bool, THREADS)
    done := make(chan bool, 1)

    // result handler
    n_requests := 1
    for i := 0; i < 4; i++ {
        n_requests *= 1 + int(scan.bytes[i][1]) - int(scan.bytes[i][0])
    }
    fmt.Println(n_requests)
    go handleResults(n_requests, queue, done)

    // launch requests
    for b0 := scan.bytes[0][0]; b0 <= scan.bytes[0][1]; b0++ {
        for b1 := scan.bytes[1][0]; b1 <= scan.bytes[1][1]; b1++ {
            for b2 := scan.bytes[2][0]; b2 <= scan.bytes[2][1]; b2++ {
                for b3 := scan.bytes[3][0]; b3 <= scan.bytes[3][1]; b3++ {
                    for _, port := range scan.ports {
                        addr := fmt.Sprintf("%d.%d.%d.%d:%d",
                                            b0, b1, b2, b3, port)
                        maxqueue <- true
                        go probe(addr, queue, maxqueue)
                    }
                }
            }
        }
    }

    // wait until all messages are handled
    <-done
}
