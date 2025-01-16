package utilu

// https://github.com/micro-services-roadmap/oneid-core/blob/master/utilo/uids.go

import (
	"crypto"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"strconv"
)

var countMap = map[crypto.Hash]int{
	crypto.MD5:    4,
	crypto.SHA1:   5,
	crypto.SHA256: 8,
	crypto.SHA512: 16,
}

var chars = []string{ // 62
	"a", "b", "c", "d", "e", "f", "g", "h",
	"i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t",
	"u", "v", "w", "x", "y", "z", "0", "1", "2", "3", "4", "5",
	"6", "7", "8", "9", "A", "B", "C", "D", "E", "F", "G", "H",
	"I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T",
	"U", "V", "W", "X", "Y", "Z",
}

func Sha512UID(key string, length int) []string {
	return GetUID(key, length, crypto.SHA512)
}

func Sha256UID(key string, length int) []string {
	return GetUID(key, length, crypto.SHA256)
}

func Sha1UID(key string, length int) []string {
	return GetUID(key, length, crypto.SHA1)
}

func Md5UID(key string, length int) []string {
	return GetUID(key, length, crypto.MD5)
}

// GetUID 产生唯一ID
// 方式1: 根据传入的值算摘要, 将摘要分段转换为10进制(index值): 自定义长度
//
// 方式2: 生成雪花ID, 然后做进制转换(62进制): 11位
//
//	refer: https://github.com/baidu/uid-generator?tab=readme-ov-file#snowflake
func GetUID(key string, length int, alg crypto.Hash) []string {
	count, hash := Calculate(key, alg)

	var hexStr = hex.EncodeToString(hash)
	resUrl := make([]string, count)

	moveBit := 32 / length
	for i := 0; i < count; i++ {
		subString := hexStr[i*8 : i*8+8] // such as
		lHexLong, _ := strconv.ParseInt(subString, 16, 64)
		outChars := ""
		for j := 0; j < length; j++ { // 使用 subString 生成一个随机串
			// 0x0000003D: 111101(61), 与之&操作结果一定小于61(chars长度是62)
			index := lHexLong & 0x0000003D
			outChars += chars[index]
			lHexLong = lHexLong >> moveBit
		}

		// 把字符串存入对应索引的输出数组
		resUrl[i] = outChars
	}

	return resUrl
}

func Calculate(key string, alg crypto.Hash) (int, []byte) {
	count := 4
	algs := crypto.MD5
	if v, ok := countMap[alg]; ok {
		count = v
		algs = alg
	}

	var hash []byte
	switch algs {
	case crypto.SHA1:
		tmp := sha1.Sum([]byte(key))
		hash = tmp[:]
	case crypto.SHA256:
		tmp := sha256.Sum256([]byte(key))
		hash = tmp[:]
	case crypto.SHA512:
		tmp := sha512.Sum512([]byte(key))
		hash = tmp[:]
	default:
		tmp := md5.Sum([]byte(key))
		hash = tmp[:]
	}
	return count, hash
}
