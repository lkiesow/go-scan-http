package main

import "net"
import "fmt"
import "bufio"
import "time"

var REQ_BEGIN []byte = []byte("HEAD / HTTP/1.1\r\nHost:")
var REQ_END []byte = []byte("\r\n\r\n")
const TIMEOUT time.Duration = 3 * time.Second
const THREADS int = 512

func probe(addr string, ret chan string) {
    d := net.Dialer{Timeout: TIMEOUT}
    conn, err := d.Dial("tcp", addr)
    if err != nil {
        ret <- fmt.Sprintf("%s", err)
        return
    }
    defer conn.Close()
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

func main() {
    queue := make(chan string, THREADS)
    queue := make(chan bool, THREADS)
    // https://ipinfo.io/AS680/131.173.0.0/16
    init := THREADS
    for i := 168; i < 168 + 26; i++ {
        for j := 0; j < 256; j++ {
            addr := fmt.Sprintf("131.173.%d.%d:80", i, j)
            //addr := fmt.Sprintf("131.173.168.%d:80", i)
            //addr := fmt.Sprintf("141.100.10.%d:80", i)
            //addr := fmt.Sprintf("52.73.210.%d:80", i)
            //addr := fmt.Sprintf("192.168.1.%d:80", i)
            go probe(addr, queue)
            go probe(addr + "80", queue)
            fmt.Println(init, i, j)
            if init > 0 {
                init--
            } else {
                result := <-queue
                fmt.Println(result)
            }
        }
    }
    for i := 0; i < THREADS * 2; i++ {
        addr := <-queue
        fmt.Println(addr)
    }
}
