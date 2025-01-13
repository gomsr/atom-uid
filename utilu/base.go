package utilu

import "strings"

const Base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func ToBase62R(n int64) string {
	if n == 0 {
		return "0"
	}

	var result strings.Builder
	for n > 0 {
		remainder := n % 62
		result.WriteByte(Base62Chars[remainder])
		n = n / 62
	}

	return result.String()
}

func ToBase62(n int64) string {
	return Reverse(ToBase62R(n))
}

func Base62ToDecimal(base62 string) int64 {
	var result int64
	for _, char := range base62 {
		value := int64(strings.Index(Base62Chars, string(char)))
		// 计算62进制的每一位对应的十进制值
		result = result*62 + value
	}
	return result
}

func DecimalToBinary(decimal int64) string {
	if decimal == 0 {
		return "0"
	}
	var result strings.Builder
	for decimal > 0 {
		remainder := decimal % 2
		result.WriteByte(byte(remainder + '0')) // 将余数转换为字符 '0' 或 '1'
		decimal /= 2
	}
	// 反转结果
	return Reverse(result.String())
}

func Reverse(s string) string {
	var reversed []rune
	for i := len(s) - 1; i >= 0; i-- {
		reversed = append(reversed, rune(s[i]))
	}
	return string(reversed)
}
