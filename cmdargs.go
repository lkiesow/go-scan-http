package main

import (
    "errors"
    "fmt"
    "os"
    "regexp"
    "strconv"
    "strings"
)

type scanrange struct {
    bytes [4][2]int
    ports []int
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
// The method returns a scanrange struct, specifying a rage of IP addresses and
// a list of ports.
func parseRangeString(args []string) (scanrange, error) {
    scan := scanrange{}

    if len(args) == 0 {
        return scan, errors.New("Not enough arguments")
    }

    re := regexp.MustCompile(`^([0-9]+)\.([0-9]+)\.([0-9]+)\.([0-9]+)/([0-9]+)$`)
    if !re.MatchString(args[0]) {
        return scan, errors.New("Invalid range argument")
    }
    parts := re.FindStringSubmatch(args[0])
    var addr uint64 = 0
    for _, part := range parts[1:5] {
        b, _ := strconv.ParseUint(part, 10, 32)
        addr <<= 8
        addr |= b
    }
    mask, _ := strconv.ParseUint(parts[5], 10, 32)
    var mask_h uint64 = (1 << (32 - mask)) - 1
    var mask_l uint64 = ((1 << 32) - 1) ^ mask_h
    low := addr & mask_l + 1
    high := addr | mask_h
    if (high & 255) == 255 {
        high--
    }
    for i := 3; i >= 0; i-- {
        scan.bytes[i][0] = int(low & 255)
        scan.bytes[i][1] = int(high & 255)
        low >>= 8
        high >>= 8
    }

    scan.ports, _ = parsePorts(args[1:])

    return scan, nil
}

// parseRangeArgs takes the command line arguments as slice of strings,
// expecting the first four arguments to be a specification of a IP addresses.
// For example:
//
//     192 168 0-1 *
//
// Additional arguments are parsed as ports.
//
// The method returns a scanrange struct, specifying a rage of IP addresses and
// a list of ports.
func parseRangeArgs(args []string) (scanrange, error) {
    scan := scanrange{}

    if len(args) < 4 {
        return scan, errors.New("Not enough arguments")
    }

    for i := 0; i < 4; i++ {
        if args[i] == "*" {
            if i < 3 {
                scan.bytes[i][0] = 0
                scan.bytes[i][1] = 255
            } else {
                scan.bytes[i][0] = 1
                scan.bytes[i][1] = 254
            }
        } else {
            byterange := strings.SplitN(args[i], "-", 2)
            val, _ := strconv.Atoi(byterange[0])
            scan.bytes[i][0] = val
            if len(byterange) == 1 {
                scan.bytes[i][1] = scan.bytes[i][0]
            } else {
                val, _ := strconv.Atoi(byterange[1])
                scan.bytes[i][1] = val
            }
        }
    }

    scan.ports, _ = parsePorts(args[4:])

    return scan, nil
}

// usage prints the usage information for go-scan-http
func usage() {
    fmt.Printf("Usage: %s addr-range | [b1 ... b4]  [ports ...]\n\n", os.Args[0])
    fmt.Println("addr-range")
    fmt.Println("  Address range specification as single string in the form")
    fmt.Println("  b1.b2.b3.b4/mask. E.g. `192.168.1.0/24`")
    fmt.Println("b[1-4]")
    fmt.Println("  Specification for a byte range to scan.")
    fmt.Println("  E.g. `10` or `1-254` or `100-150`.")
    fmt.Println("  Using the special value `*` is equivalent to `1-254`.")
    fmt.Println("ports")
    fmt.Println("  List of ports to scan.")
    fmt.Println("  This defaults to 80.")
    fmt.Println("\nExample")
    fmt.Println("  Scan a 192.168.1.0/24 network for ports 80 and 8080.")
    fmt.Println("  All these forms are equivalent.")
    fmt.Printf("  %s 192.168.1.0/24  80 8080\n", os.Args[0])
    fmt.Printf("  %s 192 168 1 1-254 80 8080\n", os.Args[0])
    fmt.Printf("  %s 192 168 1 *     80 8080\n", os.Args[0])
    os.Exit(0)
}

// parseArgs reads the command line arguments and passes them to the correct
// methods for paring the IP range information and ports.
func parseArgs() scanrange {
    args := os.Args[1:]
    scan, err := parseRangeString(args)
    if err != nil {
        scan, err = parseRangeArgs(args)
    }
    if err != nil {
        usage()
    }
    return scan
}
