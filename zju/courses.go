package zju

import (
	"net/http"
	"regexp"
	"io"
	"log"
	"fmt"
	"net/url"
	"context"
)
type Courses struct {
	am *ZJUAM
	firstTime bool
}
func NewCourses (username, password string) *Courses {
	return &Courses {
		am: NewZJUAM(username, password, WithRedirectsDisabled()),
		firstTime: true,
	}
}
// Choose the output length and truncate it
func Truncate (s string, n ...int) string {
	r := []rune(s)
	var length int
	if len(n) == 0 {
		length = 50
	} else {
		length = n[0]
	}
	if len(s) > length {
		return string(r[:length]) + "..."
	} else {
		return s
	}
}
func (c *Courses) login (ctx context.Context) {
	fmt.Println("[COURSES] login begins")
	currentURL := "https://courses.zju.edu.cn/user/index"
	URL, _ := url.Parse(currentURL)
	for URL.Host != "zjuam.zju.edu.cn" {
		fmt.Printf("[COURSES] Redirect: %s\n", Truncate(URL.String()))
		res, err := c.am.ReqOfRes(ctx, nil, URL.String())
		if err != nil {
			log.Panic(err)
		}
		defer res.Body.Close()
		currentURL = res.Header.Get("Location")
		URL, _ = url.Parse(currentURL)
	}
	fmt.Println("[COURSES] Redirected to ZJUAM for authentication: ")
	base, _ := url.Parse("https://zjuam.zju.edu.cn/cas/login")
	q := base.Query()
	q.Set("service", URL.Query().Get("service"))
	base.RawQuery = q.Encode()
	fmt.Println(Truncate(base.String(), 70))
	currentURL, err := c.am.Att(ctx, base.String())
	if err != nil {
		log.Panic(err)
	}
	// defer res.Body.Close()
	// switch res.StatusCode {
	// case 200:
	// 	log.Panic("Failed to get the Location, Status code 200")
	// case 302:
	// 	currentURL = res.Header.Get("Location")
	// 	if currentURL == "" {
	// 		log.Panic("Failed to get the Location")
	// 	}
	// default: 
	// 	log.Panic("Unknown error when getting Location")
	// }
	fmt.Println("[COURSES] Returned from ZJUAM, final login: ")
	var i int
	for {
		i++	
		fmt.Printf("[COURSES] Redirect %d:\n%s\n", i, Truncate(currentURL, 70))
		res, err := c.am.Fetch(ctx, currentURL, nil)
		if err != nil {
			log.Panic(err)
		}
		defer res.Body.Close()
		str, err := io.ReadAll(res.Body)
		if err != nil {
			log.Panic(err)
		}
		text := string(str)

		if is, _ := regexp.MatchString(`meta http-equiv="refresh"`, text); res.StatusCode == 200 && is {
			cMatch := regexp.MustCompile(`meta http-equiv="refresh" content="0;URL=([^"]+)"`)
			currentURLs := cMatch.FindStringSubmatch(text)
			if currentURLs == nil {
				log.Panic("No URLs found")
			}
			currentURL = currentURLs[1]
			continue
		} else if !(res.StatusCode >= 300 && res.StatusCode < 400) {
			break
		} else {
			currentURL = res.Header.Get("Location")
		}
	}
	fmt.Println("[COURSES] Login finished.")
}
func (c *Courses) Fetch(ctx context.Context, url string, opt *RequestOptions) (*http.Response, error) {
	if c.firstTime {
		c.login(ctx)
		c.firstTime = false
	}
	fmt.Printf("[COURSES-APPLICATION] Fetching %s ...\n", url)

	return c.am.Fetch(ctx, url, opt)
}
