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

import "testing"

func testEq(a, b []int) bool {

    // If one is nil, the other must also be nil.
    if (a == nil) != (b == nil) {
        return false;
    }

    if len(a) != len(b) {
        return false
    }

    for i := range a {
        if a[i] != b[i] {
            return false
        }
    }

    return true
}

func TestParsePorts(t *testing.T) {

    cases := []struct {
		input []string
        expected []int
	}{
		{[]string{}, []int{80}},
        {[]string{"123", "234"}, []int{123, 234}},
	}
	for _, testcase := range cases {
        result, err := parsePorts(testcase.input)
        if err != nil || !testEq(result, testcase.expected) {
            t.Errorf("Expected %v but got %v", testcase.expected, result)
        }
    }

    result, err := parsePorts([]string{"invalid"})
    if err == nil {
        t.Errorf("Expected error but got result %v", result)
    }
}

func TestParseRangeString(t *testing.T) {

    cases := []struct {
		input []string
        expected [4][2]int
	}{
		{[]string{"127.0.0.0/24"},
         [4][2]int{{127, 127}, {0, 0}, {0, 0}, {1, 254}}},
		{[]string{"127.0.0.0/24", "8081"},
         [4][2]int{{127, 127}, {0, 0}, {0, 0}, {1, 254}}},
		{[]string{"127.0.0.0/16", "8081"},
         [4][2]int{{127, 127}, {0, 0}, {0, 255}, {1, 254}}},
	}
	for _, testcase := range cases {
        result, err := parseRangeString(testcase.input)
        if err != nil || result.bytes != testcase.expected {
            t.Errorf("Expected %v but got %v", testcase.expected, result.bytes)
        }
    }

    errorcases := [][]string{
		[]string{},
		[]string{"127.0.0.0"},
		[]string{"127.0.0.0/33"},
    }
	for _, testcase := range errorcases {
        result, err := parseRangeString(testcase)
        if err == nil {
            t.Errorf("Expected error but got result %v for input %v",
                     result, testcase)
        }
    }

}

func TestParseRangeArgs(t *testing.T) {

    cases := []struct {
		input []string
        expected [4][2]int
	}{
		{[]string{"127", "0", "0", "*"},
         [4][2]int{{127, 127}, {0, 0}, {0, 0}, {1, 254}}},
		{[]string{"127", "0", "0", "*", "8080"},
         [4][2]int{{127, 127}, {0, 0}, {0, 0}, {1, 254}}},
		{[]string{"127", "0", "*", "*", "8080"},
         [4][2]int{{127, 127}, {0, 0}, {0, 255}, {1, 254}}},
		{[]string{"127", "2-3", "*", "1-254", "8080"},
         [4][2]int{{127, 127}, {2, 3}, {0, 255}, {1, 254}}},
	}
	for _, testcase := range cases {
        result, err := parseRangeArgs(testcase.input)
        if err != nil || result.bytes != testcase.expected {
            t.Errorf("Expected %v but got %v", testcase.expected, result.bytes)
        }
    }

    errorcases := [][]string{
		[]string{},
		[]string{"127", "0", "0"},
		[]string{"127", "0", "0", "300"},
		[]string{"127", "0", "0", "-3"},
		[]string{"127", "0", "a", "1-3"},
		[]string{"127", "0", "0-b", "-3"},
		[]string{"127", "0", "0", "4-3"},
    }
	for _, testcase := range errorcases {
        result, err := parseRangeArgs(testcase)
        if err == nil {
            t.Errorf("Expected error but got result %v for input %v",
                     result, testcase)
        }
    }

}
