package utilu

import (
	"fmt"
	"strconv"
	"testing"
)

func TestToBase62(t *testing.T) {

	println(ToBase62(916132832))
	println(ToBase62(32590299105))
}

func TestV2(t *testing.T) {
	// 步骤 1: 将62进制数转换为十进制
	decimalValue := Base62ToDecimal("zzzzzz")
	fmt.Printf("62进制 %s 转换为十进制: %d\n", "100000", decimalValue)

	// 步骤 2: 将十进制数转换为二进制
	println(DecimalToBinary(decimalValue))

	println(strconv.ParseInt("1111111111111111111111111111111111111111", 2, 64))
}
