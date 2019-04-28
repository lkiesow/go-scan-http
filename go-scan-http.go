/*
go-scan-http -- Fast http network scanner
Copyright (C) 2019 Lars Kiesow <lkiesow@uos.de>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

// go-scan-http scans a range of IPv4 addresses for HTTP servers by sending a
// simple, short request and listening for any kind of answer. It keeps the
// response header for further investigation.
package main

import (
	"bufio"
	"fmt"
	"net"
	"time"
)

var requestBegin = []byte("HEAD / HTTP/1.1\r\nHost:")
var requestEnd = []byte("\r\n\r\n")

const timeout time.Duration = 5 * time.Second

// probe takes a single IPv4 address and sends a simple HTTP request to this
// address. The resulting HTTP header or error is written to the ret channel.
// Upon completion, one value is removed from the maxqueue channel to indicate
// that another request can be launched.
func probe(addr string, ret chan string, maxqueue chan bool) {
	d := net.Dialer{Timeout: timeout}
	conn, err := d.Dial("tcp", addr)
	if err != nil {
		ret <- fmt.Sprintf("%s", err)
		<-maxqueue
		return
	}
	defer conn.Close()
	defer func() { <-maxqueue }()
	conn.SetDeadline(time.Now().Add(timeout))
	conn.Write(requestBegin)
	conn.Write([]byte(addr))
	conn.Write(requestEnd)
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

// handleResults reads and prints the number of results defined by num from the
// results channel and writes a single value to the done channel once it is
// finished.
func handleResults(num int, results chan string, done chan bool) {
	for i := 0; i < num; i++ {
		addr := <-results
		fmt.Println(addr)
	}
	done <- true
}

// main is the entry point for the executable.
func main() {
	settings := parseArgs()

	results := make(chan string, settings.threads)
	maxqueue := make(chan bool, settings.threads)
	done := make(chan bool, 1)

	// result handler
	nRequests := len(settings.ports)
	for i := 0; i < 4; i++ {
		nRequests *= 1 + int(settings.bytes[i][1]) - int(settings.bytes[i][0])
	}

	// initialize result handler
	go handleResults(nRequests, results, done)

	// launch requests
	for b0 := settings.bytes[0][0]; b0 <= settings.bytes[0][1]; b0++ {
		for b1 := settings.bytes[1][0]; b1 <= settings.bytes[1][1]; b1++ {
			for b2 := settings.bytes[2][0]; b2 <= settings.bytes[2][1]; b2++ {
				for b3 := settings.bytes[3][0]; b3 <= settings.bytes[3][1]; b3++ {
					for _, port := range settings.ports {
						addr := fmt.Sprintf("%d.%d.%d.%d:%d",
							b0, b1, b2, b3, port)
						maxqueue <- true
						go probe(addr, results, maxqueue)
					}
				}
			}
		}
	}

	// wait until all messages are handled
	<-done
}
