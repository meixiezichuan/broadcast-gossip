package common

import (
	"golang.org/x/exp/rand"
	"reflect"
	"strconv"
)

func IsStructEmpty(s interface{}) bool {
	// 判断是否是零值
	return reflect.DeepEqual(s, reflect.Zero(reflect.TypeOf(s)).Interface())
}

func GenerateNodeInfo() map[string]string {
	cnum := rand.Intn(100)
	cpu := "%" + strconv.Itoa(cnum)

	mnum := rand.Intn(100)
	mem := strconv.Itoa(mnum) + "MB"

	bnum := rand.Intn(100)
	ba := "%" + strconv.Itoa(bnum)

	return map[string]string{
		"Cpu":     cpu,
		"Mem":     mem,
		"Battery": ba,
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Utility function to check if a slice contains a value
func Contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}
