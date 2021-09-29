package util

import (
	"fmt"
	"github.com/parnurzeal/gorequest"
	"time"
)

func Fetch(url string) (html string, err error) {

	request := gorequest.New()

	resp, body, errs := request.Get(url).
		Set("accept","text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9").
		Set(`accept-language`,`zh-CN,zh;q=0.9,en;q=0.8`).
		Set(`referer`,`https://www.indeed.com/cmp/Enterprise-Holdings/jobs?start=750`).
		Set(`sec-ch-ua`,`"GoogleChrome";v="93","Not;ABrand";v="99","Chromium";v="93"`).
		Set(`sec-ch-ua-mobile`,`?0`).
		Set(`sec-ch-ua-platform`,`"Windows"`).
		Set(`sec-fetch-dest`,`document`).
		Set(`sec-fetch-mode`,`navigate`).
		Set(`sec-fetch-site`,`same-origin`).
		Set(`sec-fetch-user`,`?1`).
		Set(`upgrade-insecure-requests`,`1`).
		Set(`user-agent`,`Mozilla/5.0(WindowsNT10.0;Win64;x64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/93.0.4577.82Safari/537.36`).
		Timeout(30 * time.Second).
		End()

	if err = ErrAndStatus(errs, resp); err != nil {
		return "", err
	} else {
		return body, err
	}
}

func Post(index int, jobKey string) (html string, err error) {
	var status int

	defer func() {
		fmt.Println(index, jobKey, "post status: ", status)
	}()

	request := gorequest.New()

	url := "https://apis.indeed.com/jobseeker/graphql?co=US"

	postData := `{"query":"\n        {\n            jobData (jobKeys:[\"`+ jobKey + `\"]) {\n                results {\n                    job {\n                        key\n                        title\n                        description\n                        {\n                            html\n                        }\n                        indeedApply {\n                            key\n                            scopes\n                        }\n                        location {\n                            countryCode\n                        }\n                    }\n                }\n            }\n        }"}`

	resp, body, errs := request.Post(url).
		Set("Accept","application/json,text/plain,*/*").
		//Set("Accept-Encoding","gzip,deflate,br").
		Set("Accept-Language","zh-CN,zh;q=0.9,en;q=0.8").
		Set("Connection","keep-alive").
		Set("Content-Type","application/json").
		Set("Host","apis.indeed.com").
		Set("Indeed-API-Key","4cac2f3fb3b9587eb5a818c802f4bf0b89bf2c8a957c2ac82c3756ad29e2b742").
		Set("Origin","https://www.indeed.com").
		Set("Referer","https://www.indeed.com/").
		Set("sec-ch-ua",`"GoogleChrome";v="93","Not;ABrand";v="99","Chromium";v="93"`).
		Set(`sec-ch-ua-mobile`,`?0`).
		Set(`sec-ch-ua-platform`,`"Windows"`).
		Set(`Sec-Fetch-Dest`,`empty`).
		Set(`Sec-Fetch-Mode`,`cors`).
		Set(`Sec-Fetch-Site`,`same-site`).
		Set(`User-Agent`,`Mozilla/5.0(WindowsNT10.0;Win64;x64)AppleWebKit/537.36(KHTML,likeGecko)Chrome/93.0.4577.82Safari/537.36`).
		Send(postData).
		Timeout(30 * time.Second).
		End()

	status = resp.StatusCode

	if err = ErrAndStatus(errs, resp); err != nil {
		return "", err
	} else {
		return body, err
	}
}

// ErrAndStatus goRequest错误判断
func ErrAndStatus(errs []error, resp gorequest.Response) (err error) {
	if len(errs) > 0 {
		err = errs[0]
		return
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("http code: %d", resp.StatusCode)
	}

	return
}
