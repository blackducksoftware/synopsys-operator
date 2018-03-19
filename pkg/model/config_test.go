package model

import (
	"fmt"
	"testing"
)

func TestConfig(t *testing.T) {
	var strArr = []string{"example1", "example2"}
	str := generateStringFromStringArr(strArr)

	if str != "[\"example1\",\"example2\"]" {
		fmt.Printf("The final string is %s", str)
		t.Fail()
	}
}
