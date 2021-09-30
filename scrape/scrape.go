package scrape

import (
	"fmt"
	"github.com/antchfx/htmlquery"
	"github.com/panjf2000/ants"
	"github.com/tidwall/gjson"
	"html"
	"indeed_job_scraper_go/config"
	"indeed_job_scraper_go/sqlite"
	"indeed_job_scraper_go/util"
	"regexp"
	"strings"
	"sync"
)

// 工作页面配置
type JobItem struct {
	location string
	jobKey   string
}

// 工作页面配置
type Profile struct {
	i           int
	jobKey      string
	location    string
	companyName string
}

// 工作信息
type JobInfo struct {
	CompanyName string
	Title       string
	Location    string
	Description string
}

var JobChan = make(chan JobInfo, 100) // 保存采集到的工作信息

// 爬取所有工作的信息
func GetJobs(companyName string, jobItems []JobItem) {
	var wg sync.WaitGroup

	p, _ := ants.NewPoolWithFunc(1, func(profile interface{}) {
		getJob(profile)
		wg.Done()
	})
	defer p.Release()

	// Submit tasks one by one.
	for i, jobItem := range jobItems {
		wg.Add(1)

		profile := Profile{
			i:           i,
			companyName: companyName,
			location:    jobItem.location,
			jobKey:      jobItem.jobKey,
		}
		_ = p.Invoke(profile)
	}

	wg.Wait()
	fmt.Printf("finish all tasks")
}

// 爬取，提取某工作信息，发送到列队
func getJob(profile interface{}) {
	defer func() {
		if err := recover(); err != nil {
			// 打印异常，关闭资源，退出此函数
			fmt.Println("err >>>>>>>>>> ", err)
			return
		}
	}()

	p := profile.(Profile)

	if sqlite.SelectUrl(p.jobKey) {
		fmt.Println(p.i, p.jobKey, " already scraped")
		return
	}

	htmlCode, err := util.Post(p.i, p.jobKey)
	if err != nil {
		fmt.Println(p.jobKey, err)
		return
	}

	if htmlCode == "" {
		fmt.Println("htmlCode is empty")
	}

	// 检查是否正确的json字符串
	if !strings.Contains(htmlCode, `{"data":{"jobData"`) {
		fmt.Println(p.i, p.jobKey, " json is not right")
	}

	// 获取json节点
	tmp := gjson.Parse(htmlCode).Get("data").Get("jobData").Get("results.0").Get("job")

	// 获取国家，标题，和描述
	country := tmp.Get("location").Get("countryCode").String()
	title := tmp.Get("title").String()
	description := htmlClear(tmp.Get("description").Get("html").String())

	fmt.Println("- extracted", title)

	if country != "US" {
		fmt.Println("- extract", title, "the county is not US")
	}

	jobInfo := JobInfo{
		CompanyName: p.companyName,
		Title:       title,
		Location:    p.location,
		Description: description,
	}

	JobChan <- jobInfo

	config.ScrapedChan <- p.jobKey
}

// 描述里面的 html 代码清洗，转移符清洗
func htmlClear(org string) (cleared string) {
	//去除所有尖括号内的HTML代码，并换成换行符
	re, _ := regexp.Compile("\\<[\\S\\s]+?\\>")
	cleared = html.UnescapeString(re.ReplaceAllString(org, ""))

	return
}

// 爬取该公司所有工作列表
func GetListPages(url string) (jobItems []JobItem) {
	//url = "https://www.indeed.com/cmp/Bristol-Myers-Squibb/jobs"
scrapeStart:
	fmt.Println("scrape url:" + url)

	htmlCode, err := util.Fetch(url)
	if err != nil {
		fmt.Println(url, err)
		return
	}
	if htmlCode == "" {
		fmt.Println("htmlCode is empty")
	}

	newJobItems := ExtractJobKey(htmlCode)

	fmt.Println("find jobKey item numbers: ", len(newJobItems))

	jobItems = append(jobItems, newJobItems...)

	nextPage := extractNext(htmlCode)
	if nextPage != "" {
		url = "https://www.indeed.com" + nextPage
		//time.Sleep(1 * time.Second)
		goto scrapeStart
	}

	pUniqueItems := new([]JobItem)

	removeDuplicate(jobItems, pUniqueItems)

	return *pUniqueItems
}

func removeDuplicate(personList []JobItem, pUniqueItems *[]JobItem) {
	for i := range personList {
		flag := true
		for j := range *pUniqueItems {
			if personList[i].jobKey == (*pUniqueItems)[j].jobKey {
				flag = false
				break
			}
		}
		if flag {
			*pUniqueItems = append(*pUniqueItems, personList[i])
		}
	}
	return
}

func extractNext(html string) string {
	doc, err := htmlquery.Parse(strings.NewReader(html))
	if err != nil {
		return ""
	}

	tmp := htmlquery.FindOne(doc, "//a[@data-tn-element='next-page']")
	if tmp == nil {
		return ""
	}

	return htmlquery.SelectAttr(tmp, "href")
}

// 提取该页面所有的 job key 及其 location
func ExtractJobKey(html string) (jobItems []JobItem) {
	doc, err := htmlquery.Parse(strings.NewReader(html))
	if err != nil {
		return
	}

	//items := htmlquery.Find(doc, "//li[contains(@class, 'cmp-JobListItem')]")
	items := htmlquery.Find(doc, `//ul[@class="cmp-JobList-jobList"]/li`)

	if items == nil {
		fmt.Println("cant find any job items")
		return
	}

	for _, row := range items {
		jobKeyTmp := htmlquery.SelectAttr(row, "data-tn-entityid")
		locationTmp := htmlquery.FindOne(row, `//div[@class='cmp-JobListItem-subtitle']`)

		re := regexp.MustCompile(`0,([a-z0-9]+?),\d`)
		match := re.FindStringSubmatch(jobKeyTmp)

		if match == nil || locationTmp == nil {
			continue
		}

		location := htmlquery.InnerText(locationTmp)

		jobItem := JobItem{location: location, jobKey: match[1]}

		jobItems = append(jobItems, jobItem)
	}

	return
}
