package utilu

// NextPowerOfTwo 要将一个数字转换为大于或等于它的 最近的 2 的次方数
//   - 输入 5: 输出 8
//   - 输入 9: 输出 16
//   - 输入 16: 输出 16
func NextPowerOfTwo[T int | int64](n T) T {
	if n <= 0 {
		return 1
	}
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	return n + 1
}
