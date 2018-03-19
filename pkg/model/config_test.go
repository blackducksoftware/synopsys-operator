package model

import (
	"fmt"
	"testing"
)

func TestConfig(t *testing.T) {
	var strArr = []string{"example1", "example2"}
	str := generateStringFromStringArr(strArr)
	fmt.Printf("The final string is %s", str)
}
