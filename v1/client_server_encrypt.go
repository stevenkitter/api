package v1

// 客户端 服务端 加解密所使用

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"log"
	"sort"
	"strings"
	"github.com/stevenkitter/api/v1/common"
)

const (
	//ContentKey content
	ContentKey = "72d716b73924becbcc6615dc5dff95e86c8098b6" //内容区块加密
	//HeaderKey key
	HeaderKey  = "998806f0b5c6611e6da50ffef63fb1747a84edfe" // 头部区域加密
)

func testAes() {
	// AES-128。key长度：16, 24, 32 bytes 对应 AES-128, AES-192, AES-256
	key := []byte("hundsun@12345678")
	result, err := AesEncrypt([]byte("polaris@studygolang"), key)
	if err != nil {
		panic(err)
	}
	fmt.Println(base64.StdEncoding.EncodeToString(result))
	origData, err := AesDecrypt(result, key)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(origData))
}

//CheckSignature is
func CheckSignature(timestamp, nonce, sign string) (bool, string) {
	ret, msg := CheckRepetRequest(nonce)
	if !ret {
		return false, msg
	}
	tmpArr := []string{HeaderKey, timestamp, nonce}
	sort.Strings(tmpArr)
	tmpStr := strings.Join(tmpArr, "")
	actual := fmt.Sprintf("%x", sha1.Sum([]byte(tmpStr)))
	return actual == sign, "头部签名匹配错误，参数不对"
}
//CheckRepetRequest is
func CheckRepetRequest(nonce string) (bool, string) {
	non, err := common.RedisGETString(nonce)
	if non != "" {
		//重复提交了
		return false, "重复提交了,有被黑的风险！"
	}
	err = common.RedisSaveStringEx(nonce, nonce, "600")
	if err != nil {
		log.Println(err.Error())
		return false, err.Error()
	}
	return true, "ok"
}
//AesEncrypt is
func AesEncrypt(origData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	origData = PKCS5Padding(origData, blockSize)
	// origData = ZeroPadding(origData, block.BlockSize())
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypted := make([]byte, len(origData))
	// 根据CryptBlocks方法的说明，如下方式初始化crypted也可以
	// crypted := origData
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}
//AesDecrypt is
func AesDecrypt(crypted, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(crypted))
	// origData := crypted
	blockMode.CryptBlocks(origData, crypted)
	origData = PKCS5UnPadding(origData)
	// origData = ZeroUnPadding(origData)
	return origData, nil
}
//ZeroPadding is
func ZeroPadding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{0}, padding)
	return append(ciphertext, padtext...)
}
//ZeroUnPadding is
func ZeroUnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}
//PKCS5Padding is
func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}
//PKCS5UnPadding is
func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	// 去掉最后一个字节 unpadding 次
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}
