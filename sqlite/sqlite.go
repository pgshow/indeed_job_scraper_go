package sqlite

import (
	"database/sql"
	"fmt"
	"github.com/astaxie/beego/logs"
	_ "github.com/mattn/go-sqlite3"
	"indeed_job_scraper_go/config"
	"os"
	"strings"
	"sync"
)

var (
	DB = sql.DB{}
	s  sync.RWMutex
)

func DbInit() {
	DB = *openDB("db.db")
}

func openDB(dbPath string) *sql.DB {
	// 判断数据库是否存在
	_, err := os.Stat(dbPath) //os.Stat获取文件信息
	if err != nil {
		panic(err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		panic(err)
	}

	return db
}

func ScrapedSave() {
	// 爬过的项目要记录在数据库
	for {
		item := <-config.ScrapedChan
		addUrl(item)
	}
}

func addUrl(url string) bool {
	stmt, err := DB.Prepare("INSERT INTO scraped(url) values(?)")
	if err != nil {
		logs.Error("Add url to sqlite3 faild: ", err)
		return false
	}
	defer stmt.Close()

	result, err := stmt.Exec(url)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return false
		}
		logs.Error("Add url to sqlite3 faild: ", err)
		return false
	}

	if result != nil {
		return true
	}
	return false
}

func SelectUrl(url string) (exist bool) {
	sqlStr := fmt.Sprintf("SELECT id FROM 'scraped' WHERE url = '%s'", url)
	s.RLock()
	defer s.RUnlock()
	rows, err := DB.Query(sqlStr)
	if err != nil {
		logs.Error("Select url failed: ", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		err = rows.Scan(&id)
		if id > 0 {
			return true
		}
	}

	return
}
