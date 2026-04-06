package zju

import (
	"net/url"
	"math/big"
	"strings"
	"encoding/json"
	"regexp"
	"io"
	"context"
	"fmt"
	"net/http"
	"net/http/cookiejar"
)

type ZJUAM struct {
	username string
	password string
	client *http.Client
	jar http.CookieJar
	loggedIn bool
}

type RequestOptions struct {
	Method string
	Headers http.Header
	Body io.Reader
}

type PubKey struct {
	// The initials should be capitalized to be accessible
	Modulus string `json:"modulus"`
	Exponent string `json:"exponent"`
}
// Specifically for error reporting of ZJUAM
type ZJUAMError struct {
	error string
}

func (z *ZJUAMError) Error() string {
	return z.error
}

type Option func (*ZJUAM)
func WithRedirectsEnabled () Option {
	return func (z *ZJUAM) {
		z.client.CheckRedirect = nil
	}
}

func WithRedirectsDisabled () Option {
	return func (z *ZJUAM) {
		z.client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}
}
// Explicitly pass the function names into the constuctor makes it
// very clear and intuitive
func NewZJUAM (username, password string, opts ...Option) *ZJUAM {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: jar,
	}
	z := &ZJUAM {
		username: username,
		password: password,
		client: client,
		jar: jar,
		loggedIn: false,
	}
	for _, opt := range opts {
		opt(z)
	}
	return z
}

const pubkey_url string = "https://zjuam.zju.edu.cn/cas/v2/getPubKey"

func (z *ZJUAM) ReqOfRes (ctx context.Context, opt *RequestOptions, url string) (*http.Response, error) {
	var (
		Body io.Reader = nil
		Method string = "GET"
	)
	if opt != nil{
		if opt.Method != "" {
			Method = opt.Method
		}
		if opt.Body != nil {
			Body = opt.Body
		}
	}
	req, err := http.NewRequestWithContext(ctx, Method, url, Body)
	if err != nil {
		return nil, &ZJUAMError {error: "Make request failed: " + err.Error()}
	}
	if opt != nil && opt.Headers != nil {
		req.Header = opt.Headers
	}
	resp, err := z.client.Do(req)
	if err != nil {
		return nil, &ZJUAMError { error: "Failed to get http response in ReqOfRes()" }
	}
	return resp, nil
	// Only *ZJUAMError has the Error() method to satisfy the error interface
}
func (z *ZJUAM) getStream (ctx context.Context, requestOptions *RequestOptions, url string) (io.ReadCloser, error) {
	resp, err := z.ReqOfRes(ctx, requestOptions, url)
	if err != nil {
		return nil, &ZJUAMError { error: "client.Do() in getStream() failed: " + err.Error() + "\n"}
	}
	// We can't do defer resp.Body.Close() here because it will be 
	// executed before the return statement
	return resp.Body, nil
}

func convertStreamToString (stream io.Reader) string {
	str, err := io.ReadAll(stream)
	if err != nil {
		return ""
	}
	return string(str)
}

func (z *ZJUAM) getString (ctx context.Context, requestOptions *RequestOptions, url string) (string, error) {
	stream, err := z.getStream(ctx, requestOptions, url)
	// We close the io.Reader here
	if err != nil {
		return "", &ZJUAMError { error: "client.Do() in getStream() failed: " + err.Error() + "\n"}

	}
	defer stream.Close()
	// Note that this is the standard pattern to do so:
	// 1. Acquire. 2. Check error. 3. defer the cleanup
	body := convertStreamToString(stream)
	if body == "" {
		return "", &ZJUAMError { error: "There's nothing returned" + "\n" }
	}
	return body, nil
}

// You must pass an address as v, or it won't change anything
func convertStreamToJSON (stream io.Reader, v any) error {
	return json.NewDecoder(stream).Decode(v)
}
// You must pass an address as v, or it won't change anything
func (z *ZJUAM) getJSON (ctx context.Context, requestOptions *RequestOptions, url string, v any) error {
	stream, err := z.getStream(ctx, requestOptions, url)
	if err != nil {
		return &ZJUAMError { error: "getStream() in getJSON() failed: " + err.Error() + "\n" }
	}
	defer stream.Close()
	// Note that the defer will be done before back to the caller.
	// It will definitely be done after the return value is decided.
	return convertStreamToJSON(stream, v)
}
func (z *ZJUAM) login (ctx context.Context, requestOptions *RequestOptions, loginURL string) (*http.Response, error) {
	fmt.Println("[Golang-ZJUAM] Attempting to login to ZJUAM")
	login_html, err := z.getString(ctx, requestOptions, loginURL)
	if err != nil {
		return nil , &ZJUAMError { error: "Can't access loginURL" }
	}

	matchMethod := regexp.MustCompile(`name="execution" value="([^"]+)"`)
	execution := matchMethod.FindStringSubmatch(login_html)
	if execution == nil {
		return nil, &ZJUAMError { error: "First-time login page doesn't contain execution string" }
	}
	executionValue := execution[1]
	fmt.Println("Got the executionValue")
	
	var pubkey PubKey
	err = z.getJSON(ctx, requestOptions, pubkey_url, &pubkey)
	// We can do := here because err has been defined and there's no other undefined variables!
	if err != nil {
		return nil, &ZJUAMError { error: "json.Decode() for pubkey failed. " + err.Error() + "\n" }
	}

	key, err := rsaEncrypt(z.password, pubkey.Exponent, pubkey.Modulus)
	if err != nil {
		return nil, &ZJUAMError { error: "RSA Encryption failed: " + err.Error() }
	}
	fmt.Println("Got the RSA Public Key")
	// formData := url.Values{}
	// formData.Set("username", z.username)
	// formData.Set("password", key)
	// formData.Set("execution", executionValue)
	// formData.Set("_eventId", "submit")
	// formData.Set("authcode", "")
	// There is a cleaner version
	formData := url.Values {
		"username": {z.username},
		"password": {key},
		"execution": {executionValue},
		"_eventId": {"submit"},
		"authcode": {""},
	}
	encodingString := formData.Encode()
	bodyReader := strings.NewReader(encodingString)

	reqContent := &RequestOptions {
		Method: "POST",
		Body: bodyReader,
		Headers: http.Header {
			"Content-Type": []string { "application/x-www-form-urlencoded"},
			"User-Agent": []string {
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36" },
		},
	}
	res, err := z.ReqOfRes(ctx, reqContent, loginURL)
	if err != nil {
		return nil, &ZJUAMError { error: "POST to ZJUAM failed" + err.Error() }
	}
	return res, nil
}
// Get the Location field of response
func (z *ZJUAM) Login (ctx context.Context) (string, error) {
	opt := &RequestOptions {
		Headers: nil,
		Method: "GET",
		Body: nil,
	}
	res, err := z.login (ctx, opt, "https://zjuam.zju.edu.cn/cas/login")
	if err != nil {
		return "", &ZJUAMError { error: "z.login() failed: " + err.Error() }
	}
	return z.StatusCheck(res)
}
func (z *ZJUAM) StatusCheck(res *http.Response) (string, error) {
	defer res.Body.Close()
	switch res.StatusCode {
	case 302:
		z.loggedIn = true
		fmt.Println("[Golang-ZJUAM] Login Success")
		return res.Header.Get("Location"), nil
	case 200:
		fmt.Println("Redirection failed")
		return "", &ZJUAMError { error: "Failed to redirect"}
	default:
		return "", &ZJUAMError { error: "Failed to login with status code: " + res.Status }
	}
}
// Also we can only pass nil to Fetch function here, while not in ANY OTHER PLACES!!!
// Since we change the returning type to *http.Response, Callers of the Fetch function MUST CLOSE THE BODY!!!
func (z *ZJUAM) Fetch (ctx context.Context, url string, opt *RequestOptions) (*http.Response, error) {
	if opt == nil {
		opt = &RequestOptions {
			Headers: http.Header {
				"User-Agent": []string {
					"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36",
				},
			},
			Method: "GET",
			Body: nil,
		}
	}
	if !z.loggedIn {
		fmt.Println("First Login...")
		_, err := z.Login(ctx)
		if err != nil {
			return nil, &ZJUAMError { error: "Login to ZJUAM failed" }
		}
	}
	// If you have entered into ZJUAM, then you don't need to do again.
	return z.ReqOfRes(ctx, opt, url)
}
func (z *ZJUAM) Att (ctx context.Context, url string) (string, error) {
	res, err := z.login(ctx, nil, url)
	if err != nil {
		return "", &ZJUAMError { error: "Fuck you!" }
	}
	return z.StatusCheck(res)
}
func (z *ZJUAM) AttRes (ctx context.Context, url string) (*http.Response, error) {
	res, err := z.login(ctx, nil, url)
	if err != nil {
		return nil, &ZJUAMError { error: "Fuck you!" }
	}
	return res, nil
}
func rsaEncrypt (password string, exponent string, modulus string) (string, error) {
	pwd := new(big.Int).SetBytes([]byte(password))

	n, okn := new(big.Int).SetString(modulus, 16)
	e, oke := new(big.Int).SetString(exponent, 16)
	if ! (okn && oke) {
		return "", &ZJUAMError { error: "The string -> hex failed" }
	}

	crypt := new(big.Int).Exp(pwd, e, n)

	keyLen := len(modulus)
	ciphertext := fmt.Sprintf("%0*x", keyLen, crypt)

	return ciphertext, nil
}
