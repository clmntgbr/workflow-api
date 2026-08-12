package step

import (
	"strconv"
	"strings"
)

const indexBase = 100
const indexMaxDepth = 4

func CalculateExecutionOrder(index string) int {
	parts := strings.Split(index, ".")
	result := 0
	for _, part := range parts {
		val, _ := strconv.Atoi(part)
		result = result*indexBase + val
	}
	for i := len(parts); i < indexMaxDepth; i++ {
		result *= indexBase
	}
	return result
}
