package public

import (
	"crypto/sha256"
	"fmt"
)

// SaltPassword 密码加密
func SaltPassword(salt,password string) string {
	s1:=sha256.New()
	s1.Write([]byte(password))
	str:=fmt.Sprintf("%x",s1.Sum(nil))
	s2:=sha256.New()
	s2.Write([]byte(str+salt))
	return fmt.Sprintf("%x",s2.Sum(nil))
}
