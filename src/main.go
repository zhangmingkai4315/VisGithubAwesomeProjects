package main

import (
	"database/sql"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"strings"
	"utils"
)

// root_url is the root link to all awesome projects.
const (
	root_url = "https://github.com/sindresorhus/awesome"
)

const (
	NoStatus = iota //0
	Start    = iota //1
	Fetching = iota //2
	Done     = iota //3
	Error    = iota //4
)

type Repository struct {
	Url          string
	User         string
	Star         int
	Watch        int
	Fork         int
	Status       int
	ErrorMessage string
}

type Store struct {
	Topic  string
	Url    string
	Status int
}

var StatusList map[int]string

type AwesomeProjects map[string]*Store

func (a AwesomeProjects) PutTopicAndUrl(topic string, key string, url string) {
	_, ok := a[key]
	if ok == true {
		a[key].Url = url
		a[key].Topic = topic
		a[key].Status = NoStatus
	} else {
		a[key] = &Store{Url: url, Topic: topic, Status: NoStatus}
	}
	return
}

func (a AwesomeProjects) SaveToDatabase(Db *sql.DB) (err error) {

	for k, v := range a {
		status, ok := StatusList[v.Status]
		if !ok {
			status = StatusList[0]
		}
		Db.Exec("INSERT INTO Category(name, url, topic, status) VALUES(?, ?, ?, ?) ON DUPLICATE KEY UPDATE url=?, topic=?, status=?",
			k, v.Url, v.Topic, status, v.Url, v.Topic, status)
		if err != nil {
			log.Println(err.Error())
			continue
		}
	}
	return
}
func (a AwesomeProjects) String() (s string) {
	//var temp =""
	for k, v := range a {
		status, ok := StatusList[v.Status]
		if !ok {
			status = StatusList[0]
		}
		s += fmt.Sprintf("%s : {url:%s,fetched:%v,topic:%s}\n", k, v.Url, status, v.Topic)
		fmt.Printf("%+v %+v", k, v)
	}
	return s
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
func getTitleSelections(doc *goquery.Document, selections []*goquery.Selection) []*goquery.Selection {
	var titleFilter = []string{"Contents", "License"}
	doc.Find("h2").Each(func(i int, s *goquery.Selection) {
		//we need filter to remove the useless h2 title.
		if !contains(titleFilter, s.Text()) && strings.TrimSpace(s.Text()) != "" {
			selections = append(selections, s)
		}
	})
	return selections
}
func getSelectionsLinks(selections []*goquery.Selection, awesome *AwesomeProjects) {
	for i := 0; i < len(selections); i++ {
		topic := selections[i].Text()
		selections[i].Next().Find("ul li a").Each(func(j int, s *goquery.Selection) {
			if href, exists := s.Attr("href"); exists {
				awesome.PutTopicAndUrl(topic, s.Text(), href)
			}
		})
	}
}

func init() {
	StatusList = map[int]string{
		0: "NoStatus",
		1: "Start",
		2: "Fetching",
		3: "Done",
		4: "Error",
	}
}

func main() {
	var awsomeProjects AwesomeProjects
	Db := utils.GetDBHandler()
	defer Db.Close()
	err := Db.Ping()
	utils.CheckErrorPanic(err)
	awsomeProjects = make(map[string]*Store)
	titleSelections := []*goquery.Selection{}
	doc, err := goquery.NewDocument(root_url)
	utils.CheckErrorPanic(err)
	titleSelections = getTitleSelections(doc, titleSelections)
	getSelectionsLinks(titleSelections, &awsomeProjects)
	err = awsomeProjects.SaveToDatabase(Db)
	utils.CheckErrorPanic(err)
}
