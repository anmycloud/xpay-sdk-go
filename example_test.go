package xpay

import (
	"fmt"
	"net/http"
)

func Example() {
	client, err := NewClient(
		"your platform code",
		"xpay's gateway",
		"your app private key path",
		"xpay's public key path",
	)
	if err != nil {
		fmt.Printf("%v", err)
	}

	//qr pay
	respQr, err := client.QrPay(&QrPayRequest{
		Merchant:       "123456@gmail.com",
		TotalAmount:    100,
		TradeNo:        "unique trade_no in your app",
		ProductCode:    ProductCodeAlipayQr,
		NotifyUrl:      "http://your.host.com",
		Subject:        "online trade",
		Body:           "pay for commodity",
		BusinessParams: "this message will be returned in notification data",
	})
	fmt.Println(respQr, err)

	//query
	respQuery, err := client.Query(&QueryRequest{TradeNo: "your trade_no"})
	fmt.Println(respQuery, err)

	//notification data
	var notificationHandler http.HandlerFunc
	notificationHandler = func(writer http.ResponseWriter, request *http.Request) {
		order := OrderItem{}
		if err := client.ParseNotification(request, &order); err != nil {
			fmt.Println(err)
			// handle error
			// do something
			return
		}
		fmt.Println(order)
		// do something
	}
	// add handler in your http server
	// for example,
	http.HandleFunc("/", notificationHandler)
}
