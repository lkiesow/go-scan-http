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
