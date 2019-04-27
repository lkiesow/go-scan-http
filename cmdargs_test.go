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
        result, err := parseRangeString(&settings{}, testcase.input)
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
        result, err := parseRangeString(&settings{}, testcase)
        if err == nil {
            t.Errorf("Expected error but got result %v for input %v",
                     result, testcase)
        }
    }

    // test that settings are properly modified
    inputSettings := settings{}
    inputArgs := []string{"127.0.0.0/24"}
    parseRangeString(&inputSettings, inputArgs)
    if inputSettings.bytes[0][0] != 127 {
        t.Errorf("Expected input settings to be modifed")
    }

}

func TestParseRangeArgs(t *testing.T) {

    // test parsing valid address ranges
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
        result, err := parseRangeArgs(&settings{}, testcase.input)
        if err != nil || result.bytes != testcase.expected {
            t.Errorf("Expected %v but got %v", testcase.expected, result.bytes)
        }
    }

    // test invalid address ranges
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
        result, err := parseRangeArgs(&settings{}, testcase)
        if err == nil {
            t.Errorf("Expected error but got result %v for input %v",
                     result, testcase)
        }
    }

    // test that settings are properly modified
    inputSettings := settings{}
    inputArgs := []string{"127", "0", "0", "1"}
    parseRangeArgs(&inputSettings, inputArgs)
    if inputSettings.bytes[0][0] != 127 {
        t.Errorf("Expected input settings to be modifed")
    }

}

func TestParseOptions(t *testing.T) {

    // test success
    cases := []struct {
        input []string
        threads int
        length int
    }{
        {[]string{"-t", "123", "127.0.0.1/24"}, 123, 1},
        {[]string{"127.0.0.1/24"}, 0, 1},
    }
    for _, testcase := range cases {
        settings, args, err := parseOptions(&settings{}, testcase.input)
        if err != nil || settings.threads != testcase.threads || len(args) != testcase.length {
            t.Errorf("Got %v and %v", settings, args)
        }
    }

    // test invalid input
    errorcases := []struct {
        input []string
        threads int
        length int
    }{
        {[]string{"-t", "127.0.0.1/24"}, 123, 1},
        {[]string{"-x", "127.0.0.1/24"}, 0, 1},
    }
    for _, testcase := range errorcases {
        settings, args, err := parseOptions(&settings{}, testcase.input)
        if err == nil {
            t.Errorf("Expected error but got result %v, args: %v for input %v",
                     settings, args, testcase)
        }
    }

    // test settings are properly modified
    inputSettings := settings{}
    inputArgs := []string{"-t", "123", "127.0.0.1/24"}
    parseOptions(&inputSettings, inputArgs)
    if inputSettings.threads != 123 {
        t.Errorf("Expected input settings to be modifed")
    }

}
