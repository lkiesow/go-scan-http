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

package main

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type settings struct {
	bytes   [4][2]int
	ports   []int
	threads int
}

// parsePorts takes a slice of string command line arguments and converts those
// to a slice of ports represented as integers. If no arguments are passed to
// this method, a default of port 80 is returned.
func parsePorts(args []string) ([]int, error) {
	if len(args) == 0 {
		return []int{80}, nil
	}
	ports := make([]int, len(args))
	for i, port := range args {
		iport, err := strconv.Atoi(port)
		if err != nil {
			return []int{}, err
		}
		ports[i] = iport
	}
	return ports, nil
}

// parseRangeString takes the command line arguments as slice of strings,
// expecting the first argument to be a specification of an IP address range in
// the form b0.b1.b2.b3/mask (e.g. 192.168.1.0/24).
//
// Additional arguments are parsed as ports.
//
// The method returns a settings struct, specifying a rage of IP addresses and
// a list of ports.
func parseRangeString(settings *settings, args []string) (*settings, error) {
	if len(args) == 0 {
		return settings, errors.New("Not enough arguments")
	}

	re := regexp.MustCompile(`^([0-9]+)\.([0-9]+)\.([0-9]+)\.([0-9]+)/([0-9]+)$`)
	if !re.MatchString(args[0]) {
		return settings, errors.New("Invalid range argument")
	}
	parts := re.FindStringSubmatch(args[0])
	var addr uint64
	for _, part := range parts[1:5] {
		b, _ := strconv.ParseUint(part, 10, 32)
		addr <<= 8
		addr |= b
	}
	mask, _ := strconv.ParseUint(parts[5], 10, 32)
	if mask >= 32 {
		return settings, errors.New("Mask mmust be < 32")
	}
	var maskH uint64 = (1 << (32 - mask)) - 1
	var maskL uint64 = ((1 << 32) - 1) ^ maskH
	low := addr&maskL + 1
	high := addr | maskH
	if (high & 255) == 255 {
		high--
	}
	for i := 3; i >= 0; i-- {
		settings.bytes[i][0] = int(low & 255)
		settings.bytes[i][1] = int(high & 255)
		low >>= 8
		high >>= 8
	}

	settings.ports, _ = parsePorts(args[1:])

	return settings, nil
}

// parseRangeArgs takes the command line arguments as slice of strings,
// expecting the first four arguments to be a specification of a IP addresses.
// For example:
//
//     192 168 0-1 *
//
// Additional arguments are parsed as ports.
//
// The method returns a settings struct, specifying a rage of IP addresses and
// a list of ports.
func parseRangeArgs(settings *settings, args []string) (*settings, error) {
	if len(args) < 4 {
		return settings, errors.New("Not enough arguments")
	}

	for i := 0; i < 4; i++ {
		if args[i] == "*" {
			if i < 3 {
				settings.bytes[i][0] = 0
				settings.bytes[i][1] = 255
			} else {
				settings.bytes[i][0] = 1
				settings.bytes[i][1] = 254
			}
		} else {
			byterange := strings.SplitN(args[i], "-", 2)
			val, err := strconv.Atoi(byterange[0])
			if err != nil {
				return settings, fmt.Errorf("Invalid port number %s", byterange[0])
			}
			settings.bytes[i][0] = val
			if len(byterange) == 1 {
				settings.bytes[i][1] = settings.bytes[i][0]
			} else {
				val, err := strconv.Atoi(byterange[1])
				if err != nil {
					return settings, fmt.Errorf("Invalid port number %s", byterange[0])
				}
				settings.bytes[i][1] = val
			}
			if settings.bytes[i][1] > 255 || settings.bytes[i][0] > settings.bytes[i][1] {
				return settings, errors.New("Invalid port number")
			}
		}
	}

	settings.ports, _ = parsePorts(args[4:])

	return settings, nil
}

// usage prints the usage information for go-scan-http
func usage() {
	fmt.Printf("Usage: %s [options] addr-range | [b1 ... b4]  [ports ...]\n\n",
		os.Args[0])
	fmt.Println("go-scan-http -- Fast http network scanner")
	fmt.Println("\naddr-range")
	fmt.Println("  Address range specification as single string in the form")
	fmt.Println("  b1.b2.b3.b4/mask. E.g. `192.168.1.0/24`")
	fmt.Println("b[1-4]")
	fmt.Println("  Specification for a byte range to scan.")
	fmt.Println("  E.g. `10` or `1-254` or `100-150`.")
	fmt.Println("  Using the special value `*` is equivalent to `1-254`.")
	fmt.Println("ports")
	fmt.Println("  List of ports to scan.")
	fmt.Println("  This defaults to 80.")
	fmt.Println("\nOptions")
	fmt.Println("-n <number of threads>")
	fmt.Println("  The number of parallel requests (default: 512)")
	fmt.Println("\nExample")
	fmt.Println("  Scan a 192.168.1.0/24 network for ports 80 and 8080.")
	fmt.Println("  All these forms are equivalent.")
	fmt.Printf("  %s 192.168.1.0/24  80 8080\n", os.Args[0])
	fmt.Printf("  %s 192 168 1 1-254 80 8080\n", os.Args[0])
	fmt.Printf("  %s 192 168 1 *     80 8080\n", os.Args[0])
	os.Exit(0)
}

// parseOptions extracts known options from the given command line arguments.
func parseOptions(settings *settings, args []string) (*settings, []string, error) {
	for len(args) > 1 && len(args[0]) > 1 && args[0][0] == '-' {
		switch args[0] {
		case "-n":
			threads, err := strconv.Atoi(args[1])
			if err != nil {
				return settings, args, err
			}
			settings.threads = threads
		default:
			return settings, args, errors.New("Unknown option " + args[0])
		}
		args = args[2:]
	}
	return settings, args, nil
}

// parseArgs reads the command line arguments and passes them to the correct
// methods for paring the IP range information and ports.
func parseArgs() settings {
	args := os.Args[1:]
	settings := settings{threads: 512}

	// parse options
	_, args, err := parseOptions(&settings, args)
	if err != nil {
		usage()
	}

	_, err = parseRangeString(&settings, args)
	if err != nil {
		_, err = parseRangeArgs(&settings, args)
	}
	if err != nil {
		usage()
	}
	return settings
}
