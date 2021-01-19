package xpay

import (
	"strconv"
)

type Signable interface {
	ToStringMap() map[string]string
}

type Request struct {
	RequestNo    string `json:"request_no"`
	PlatformCode string `json:"platform_code"`
	Timestamp    int64  `json:"timestamp"`
	Sign         string `json:"sign"`
	Content      string `json:"content"`
}

func (c *Request) ToStringMap() map[string]string {
	return map[string]string{
		"request_no":    c.RequestNo,
		"platform_code": c.PlatformCode,
		"timestamp":     strconv.FormatInt(c.Timestamp, 10),
		"sign":          c.Sign,
		"content":       c.Content,
	}
}

type Container struct {
	Timestamp int64  `json:"timestamp"`
	Sign      string `json:"sign"`
	Code      string `json:"code"`
	Msg       string `json:"msg"`
	Content   string `json:"content,omitempty"`
}

func (c *Container) ToStringMap() map[string]string {
	return map[string]string{
		"timestamp": strconv.FormatInt(c.Timestamp, 10),
		"code":      c.Code,
		"msg":       c.Msg,
		"sign":      c.Sign,
		"content":   c.Content,
	}
}

type QrPayRequest struct {
	Merchant       string `json:"merchant"`                  //收款商户账号
	TotalAmount    int64  `json:"total_amount"`              //订单金额 分
	Currency       string `json:"currency"`                  //币种
	TradeNo        string `json:"trade_no"`                  //订单号
	ProductCode    string `json:"product_code"`              //产品编号
	NotifyUrl      string `json:"notify_url"`                //回调通知url
	Subject        string `json:"subject"`                   //订单主体
	Body           string `json:"body"`                      //订单描述
	BusinessParams string `json:"business_params,omitempty"` //业务数据 回调时原样返回 不超过500字
}

type QrPayResponse struct {
	PayUrl string `json:"pay_url"`
	ImgUrl string `json:"img_url"`
}

type QueryRequest struct {
	TradeNo string `json:"trade_no" binding:"required"`
}

type OrderItem struct {
	Merchant       string `json:"merchant"`         //收款商户账号
	OrderNo        string `json:"order_no"`         //订单号
	PlatformCode   string `json:"platform_code"`    //业务平台编号
	OutTradeNo     string `json:"out_trade_no"`     //业务平台单号
	RequestNo      string `json:"request_no"`       //请求编号
	ProductCode    string `json:"product_code"`     //产品编号
	BusinessParams string `json:"business_params"`  //业务数据 原样返回
	OrderType      uint8  `json:"order_type"`       //订单类型 1支付订单 2退款订单
	CreatedAt      int64  `json:"created_at"`       //创建时间 时间戳 秒
	FinishTime     int64  `json:"finish_time"`      //完成时间 时间戳 秒
	TotalAmount    int    `json:"total_amount"`     //订单金额
	Status         uint8  `json:"status"`           //订单状态 1完成 2未完成
	Subject        string `json:"subject"`          //订单主题
	Body           string `json:"body"`             //订单描述
	ChannelOrderNo string `json:"channel_order_no"` //支付通道单号
	Progress       string `json:"progress"`         //支付进度
	SourceOrderNo  string `json:"source_order_no"`  //退款的源订单号
}
