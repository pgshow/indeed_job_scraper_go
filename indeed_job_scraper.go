package main

import (
	"crypto/sha256"
	_ "crypto/sha256"
	"fmt"
	"github.com/xuri/excelize/v2"
	"indeed_job_scraper_go/config"
	"indeed_job_scraper_go/save"
	"indeed_job_scraper_go/scrape"
	"indeed_job_scraper_go/sqlite"
	"os"
)

func main() {
	sqlite.DbInit()

	go save.Save()
	go sqlite.ScrapedSave()

	rows := excelData()

	for i, row := range rows {
		if i == 0 {
			//skip the title
			continue
		}

		companyName := row[0]
		jobsPageUrl := row[3]

		sum := sha256.Sum256([]byte(jobsPageUrl))
		jobsPageUrlHash := fmt.Sprintf("%x", sum)

		if sqlite.SelectUrl(jobsPageUrlHash) {
			fmt.Println(companyName, jobsPageUrl, "already scraped")
			continue
		}

		jobKeyLocation := scrape.GetListPages(jobsPageUrl)

		fmt.Println(companyName, jobsPageUrl, "find job number:", len(jobKeyLocation))

		fmt.Println(i, companyName, "scrape all the jobs")
		scrape.GetJobs(companyName, jobKeyLocation)

		config.ScrapedChan <- jobsPageUrlHash
	}
}

func excelData() (data [][]string) {
	xlsx, err := excelize.OpenFile("./CompanyURLSendback.xlsx")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	rows, err := xlsx.GetRows("Sheet1")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return rows
}
