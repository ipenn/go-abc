package abc

import (
	"crypto/aes"
	"encoding/hex"
	"github.com/chenqgp/abc/conf"
)

func AESEncrypt(src []byte) (encrypted []byte) {
	key := conf.G2codeKey
	cipher, _ := aes.NewCipher(generateKey([]byte(key)))
	length := (len(src) + aes.BlockSize) / aes.BlockSize
	plain := make([]byte, length*aes.BlockSize)
	copy(plain, src)
	pad := byte(len(plain) - len(src))
	for i := len(src); i < len(plain); i++ {
		plain[i] = pad
	}
	encrypted = make([]byte, len(plain))
	// 分组分块加密
	for bs, be := 0, cipher.BlockSize(); bs <= len(src); bs, be = bs+cipher.BlockSize(), be+cipher.BlockSize() {
		cipher.Encrypt(encrypted[bs:be], plain[bs:be])
	}

	return encrypted
}

func AESDecrypt(encrypted []byte) (decrypted []byte) {
	key := conf.G2codeKey
	cipher, _ := aes.NewCipher(generateKey([]byte(key)))
	decrypted = make([]byte, len(encrypted))
	//
	for bs, be := 0, cipher.BlockSize(); bs < len(encrypted); bs, be = bs+cipher.BlockSize(), be+cipher.BlockSize() {
		cipher.Decrypt(decrypted[bs:be], encrypted[bs:be])
	}

	trim := 0
	if len(decrypted) > 0 {
		trim = len(decrypted) - int(decrypted[len(decrypted)-1])
	}

	return decrypted[:trim]
}

func generateKey(key []byte) (genKey []byte) {
	genKey = make([]byte, 16)
	copy(genKey, key)
	for i := 16; i < len(key); {
		for j := 0; j < 16 && i < len(key); j, i = j+1, i+1 {
			genKey[j] ^= key[i]
		}
	}
	return genKey
}

func BindOrNot(uid int) UserG2code {
	var g UserG2code
	db.Debug().Where("user_id = ? and status = 1", uid).First(&g)
	if g.Id > 0 {
		b, _ := hex.DecodeString(g.Secret)
		v := AESDecrypt(b)
		g.Secret = string(v)
	}
	return g
}

func CreateSecret(secret string, uid int) {
	g := UserG2code{
		UserId:     uid,
		Secret:     secret,
		CreateTime: FormatNow(),
	}
	v1 := AESEncrypt([]byte(g.Secret))
	g.Secret = hex.EncodeToString(v1)
	db.Debug().Create(&g)
}

func DelGoogleSecret(uid int) {
	db.Debug().Where("user_id = ? and status = 0", uid).Delete(&UserG2code{})
}

func GetUserG2(uid int) UserG2code {
	var g UserG2code
	db.Debug().Where("user_id = ?", uid).Order("create_time DESC").First(&g)
	if g.Id > 0 {
		b, _ := hex.DecodeString(g.Secret)
		v := AESDecrypt(b)
		g.Secret = string(v)
	}
	return g
}
