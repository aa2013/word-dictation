package service

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"log"
	"net/smtp"
	"strings"
	"sync"
	"time"

	"worddictation/server/global"
	"worddictation/server/util"
)

const (
	codeTTL          = 5 * time.Minute
	codeResendAfter  = time.Minute
	kindRegister     = "register"
	kindResetPwd     = "reset-pwd"
)

type codeEntry struct {
	code     string
	expireAt time.Time
	sentAt   time.Time
}

var (
	codeMu     sync.Mutex
	codeMap    = map[string]codeEntry{}
)

// SendRegisterCodeMail 发送注册验证码；1 分钟内重复请求会重发上一次的验证码。
func SendRegisterCodeMail(email string) error {
	return sendCode(kindRegister, email, "单词本注册验证码")
}

func SendResetPwdCodeMail(email string) error {
	return sendCode(kindResetPwd, email, "密码找回验证码")
}

func VerifyRegisterCode(email, code string) bool {
	return verifyCode(kindRegister, email, code)
}

func VerifyResetPwdCode(email, code string) bool {
	return verifyCode(kindResetPwd, email, code)
}

// GetCode 测试辅助：直接取当前生效的验证码
func GetCode(kind, email string) string {
	codeMu.Lock()
	defer codeMu.Unlock()
	e, ok := codeMap[kind+"-"+email]
	if !ok || time.Now().After(e.expireAt) {
		return ""
	}
	return e.code
}

func sendCode(kind, email, subject string) error {
	codeMu.Lock()
	defer codeMu.Unlock()
	key := kind + "-" + email
	if e, ok := codeMap[key]; ok && time.Now().Before(e.expireAt) {
		if time.Since(e.sentAt) < codeResendAfter {
			// 冷却期内重发旧验证码
			return sendEmail(email, subject, codeText(e.code))
		}
		code := util.RandomDigits(6)
		if err := sendEmail(email, subject, codeText(code)); err != nil {
			return err
		}
		codeMap[key] = codeEntry{code: code, expireAt: time.Now().Add(codeTTL), sentAt: time.Now()}
		return nil
	}
	code := util.RandomDigits(6)
	if err := sendEmail(email, subject, codeText(code)); err != nil {
		return err
	}
	codeMap[key] = codeEntry{code: code, expireAt: time.Now().Add(codeTTL), sentAt: time.Now()}
	return nil
}

func verifyCode(kind, email, code string) bool {
	codeMu.Lock()
	defer codeMu.Unlock()
	key := kind + "-" + email
	e, ok := codeMap[key]
	if !ok || time.Now().After(e.expireAt) || e.code != code {
		return false
	}
	delete(codeMap, key)
	return true
}

func codeText(code string) string {
	return fmt.Sprintf("你的验证码是：%s。五分钟内有效。", code)
}

func sendEmail(to, subject, text string) error {
	cfg := global.Config.Email
	if cfg.Debug {
		log.Printf("[email:debug] to=%s subject=%q text=%q", to, subject, text)
		return nil
	}
	if cfg.Host == "" || cfg.Username == "" {
		return util.Internal("SMTP 未配置，请在 config.yaml 中填写邮箱配置或开启 email.debug")
	}
	from := cfg.Sender
	if from == "" {
		from = cfg.Username
	}
	msg := buildMessage(from, to, subject, text)
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	if cfg.Port == 465 {
		return sendMailImplicitTLS(addr, cfg.Host, auth, from, to, msg)
	}
	return smtp.SendMail(addr, auth, from, []string{to}, msg)
}

func buildMessage(from, to, subject, text string) []byte {
	b64 := base64.StdEncoding.EncodeToString([]byte(subject))
	return []byte(strings.Join([]string{
		"From: " + from,
		"To: " + to,
		"Subject: =?UTF-8?B?" + b64 + "?=",
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		text,
	}, "\r\n"))
}

func sendMailImplicitTLS(addr, host string, auth smtp.Auth, from, to string, msg []byte) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: host})
	if err != nil {
		return err
	}
	defer conn.Close()
	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer c.Close()
	if ok, _ := c.Extension("AUTH"); ok {
		if err = c.Auth(auth); err != nil {
			return err
		}
	}
	if err = c.Mail(from); err != nil {
		return err
	}
	if err = c.Rcpt(to); err != nil {
		return err
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err = w.Write(msg); err != nil {
		return err
	}
	if err = w.Close(); err != nil {
		return err
	}
	return c.Quit()
}
