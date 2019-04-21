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