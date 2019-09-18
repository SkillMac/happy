package hCommon

import (
	"custom/happy/hLog"
	"github.com/go-gomail/gomail"
	"strings"
)

type EmailParam struct {
	// ServerHost 邮箱服务器地址，如腾讯企业邮箱为smtp.exmail.qq.com
	ServerHost string
	// ServerPort 邮箱服务器端口，如腾讯企业邮箱为465
	ServerPort int
	// FromEmail　发件人邮箱地址
	FromEmail string
	// FromPasswd 发件人邮箱密码（注意，这里是明文形式）
	FromPasswd string
	// Toers 接收者邮件，如有多个，则以英文逗号(“,”)隔开，不能为空
	Toers string
	// CCers 抄送者邮件，如有多个，则以英文逗号(“,”)隔开，可以为空
	CCers string
}

type AutoEmail struct {
	ep *EmailParam
	m  *gomail.Message
}

var gAutoEmail *AutoEmail

func NewAutoEmail(ep *EmailParam) *AutoEmail {

	if gAutoEmail == nil {
		hLog.Info("Init Email")
		gAutoEmail = &AutoEmail{
			ep: ep,
			m:  gomail.NewMessage(),
		}
		gAutoEmail.Init()
	}
	return gAutoEmail
}

func SendEmail(subject, body string) {
	if gAutoEmail == nil {
		return
	}
	gAutoEmail.SendEmail(subject, body)
}

func (this *AutoEmail) Init() {
	if len(this.ep.Toers) == 0 {
		return
	}
	toers := []string{}

	for _, tmp := range strings.Split(this.ep.Toers, ",") {
		toers = append(toers, strings.TrimSpace(tmp))
	}

	this.m.SetHeader("To", toers...)
	toers = []string{}

	if len(this.ep.CCers) != 0 {
		for _, tmp := range strings.Split(this.ep.CCers, ",") {
			toers = append(toers, strings.TrimSpace(tmp))
		}
		this.m.SetHeader("Cc", toers...)
	}

	this.m.SetAddressHeader("From", this.ep.FromEmail, "小魏")
}

func (this *AutoEmail) SendEmail(subject, body string) {
	this.m.SetHeader("Subject", subject)
	this.m.SetBody("text/html", body)

	d := gomail.NewDialer(this.ep.ServerHost, this.ep.ServerPort, this.ep.FromEmail, this.ep.FromPasswd)

	err := d.DialAndSend(this.m)
	if err != nil {
		hLog.Error("邮件发送失败 \n subject ===> ", subject, "\n body ===> ", body, "\n [Error] ===> ", err)
	}
}
