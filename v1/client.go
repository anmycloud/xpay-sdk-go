package xpay

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/satori/go.uuid"
)

const (
	ProductCodeWechatQr      = "1001"
	ProductCodeAlipayQr      = "2001"
	ProductCodeIntegrationQr = "4001"
)

type Client struct {
	platformCode  string
	gateway       string
	appPrivateKey *rsa.PrivateKey
	xpayPublicKey *rsa.PublicKey
	httpClient    *http.Client
}

func NewClient(platformCode, gateway, appPrivateKeyPath, xpayPublicKeyPath string) (*Client, error) {
	appPrivateKey, err := loadAppPrivateKey(appPrivateKeyPath)
	if err != nil {
		return nil, err
	}
	xpayPublicKey, err := loadXpayPublicKey(xpayPublicKeyPath)
	if err != nil {
		return nil, err
	}
	return &Client{
		platformCode:  platformCode,
		gateway:       gateway,
		appPrivateKey: appPrivateKey,
		xpayPublicKey: xpayPublicKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

// qr-code trade
func (c *Client) QrPay(req *QrPayRequest) (*QrPayResponse, error) {
	r := QrPayResponse{}
	if err := c.request(req, c.gateway+"/pay/qr", &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// trade query
func (c *Client) Query(req *QueryRequest) (*OrderItem, error) {
	r := OrderItem{}
	if err := c.request(req, c.gateway+"/pay/query", &r); err != nil {
		return nil, err
	}
	return &r, nil
}

func (c *Client) ParseNotification(r *http.Request, data interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	container := Container{}
	if err := json.Unmarshal(body, &container); err != nil {
		return err
	}
	if !c.checkSign(&container) {
		return errors.New("sign error")
	}
	if container.Code != "0000" {
		return fmt.Errorf("ERROR:%s , %s", container.Code, container.Msg)
	}
	if err := json.Unmarshal([]byte(container.Content), data); err != nil {
		return fmt.Errorf("%v,(data:%s)", err, container.Content)
	}
	return nil
}

func (c *Client) checkSign(s Signable) bool {
	params := s.ToStringMap()
	if params["sign"] == "" {
		return false
	}
	sign, err := base64.StdEncoding.DecodeString(params["sign"])
	if err != nil {
		return false
	}
	delete(params, "sign")
	signCode := mapToSignString(params)
	err = rsa.VerifyPKCS1v15(c.xpayPublicKey, crypto.SHA1, sha1([]byte(signCode)), sign)
	if err != nil {
		return false
	}
	return true
}

func (c *Client) createSign(s Signable) string {
	params := s.ToStringMap()
	if _, ok := params["sign"]; ok {
		delete(params, "sign")
	}
	signString := mapToSignString(params)
	sign, err := rsa.SignPKCS1v15(rand.Reader, c.appPrivateKey, crypto.SHA1, sha1([]byte(signString)))
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(sign)
}

func (c *Client) createRequestBody(data interface{}) (io.Reader, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	req := &Request{
		RequestNo:    uuid.NewV4().String(),
		PlatformCode: c.platformCode,
		Timestamp:    time.Now().Unix(),
		Content:      string(dataBytes),
	}
	req.Sign = c.createSign(req)
	jsonBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(jsonBytes), nil
}

func (c *Client) request(data interface{}, url string, result interface{}) error {
	body, err := c.createRequestBody(data)
	if err != nil {
		return err
	}
	request, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("response(%d):%s", resp.StatusCode, string(respBody))
	}
	res := Container{}
	if err := json.Unmarshal(respBody, &res); err != nil {
		return err
	}
	if !c.checkSign(&res) {
		return errors.New("response data sign error")
	}
	if res.Code != "0000" {
		return fmt.Errorf("ERROR:%s , %s", res.Code, res.Msg)
	}

	if err := json.Unmarshal([]byte(res.Content), result); err != nil {
		return fmt.Errorf("%v,(data:%s)", err, res.Content)
	}
	return nil
}

func loadAppPrivateKey(appPrivateKeyPath string) (*rsa.PrivateKey, error) {
	pemData, err := ioutil.ReadFile(appPrivateKeyPath)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(pemData)
	private, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return private.(*rsa.PrivateKey), nil
}

func loadXpayPublicKey(xpayPublicKeyPath string) (*rsa.PublicKey, error) {
	pemData, err := ioutil.ReadFile(xpayPublicKeyPath)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(pemData)
	public, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return public.(*rsa.PublicKey), nil
}

func mapToSignString(params map[string]string) string {
	// key 排序
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	// 拼接签名串
	buff := new(bytes.Buffer)
	for _, k := range keys {
		v := strings.TrimSpace(params[k])
		if v != "" {
			buff.WriteString(fmt.Sprintf("%s=%s&", k, v))
		}
	}
	return strings.Trim(buff.String(), "&")
}

func sha1(input []byte) []byte {
	h := crypto.SHA1.New()
	h.Write(input)
	return h.Sum(nil)
}
