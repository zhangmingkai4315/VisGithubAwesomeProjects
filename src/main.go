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
	url          string
	name         string
	topic        string
	star         int
	watch        int
	fork         int
	status       int
	errorMessage string
}

type ReposListInterface interface {
	InsertRepo(topic string, name string, url string)
}
type ReposList []Repository

func (rl ReposList) InsertRepo(topic string, name string, url string) {
	rl = append(rl, Repository{url: url, name: name, topic: topic, star: 0, watch: 0, fork: 0, status: 0, errorMessage: ""})
}

type ReposQuery struct {
	tokens    chan (struct{})
	reposList ReposList
}

type Store struct {
	topic  string
	url    string
	status int
}

var StatusList map[int]string

type AwesomeProjects map[string]*Store

func (a AwesomeProjects) PutTopicAndUrl(topic string, key string, url string) {
	_, ok := a[key]
	if ok == true {
		a[key].url = url
		a[key].topic = topic
		a[key].status = NoStatus
	} else {
		a[key] = &Store{url: url, topic: topic, status: NoStatus}
	}
	return
}

func (a AwesomeProjects) SaveToDatabase(Db *sql.DB) (err error) {

	for k, v := range a {
		status, ok := StatusList[v.status]
		if !ok {
			status = StatusList[0]
		}
		Db.Exec("INSERT INTO Category(name, url, topic, status) VALUES(?, ?, ?, ?) ON DUPLICATE KEY UPDATE url=?, topic=?, status=?",
			k, v.url, v.topic, status, v.url, v.topic, status)
		if err != nil {
			log.Println(err.Error())
			continue
		}
	}
	return
}

func (a AwesomeProjects) GetRepos(Db *sql.DB, rq *ReposQuery) {
	queue := make(chan ReposList)
	for k, v := range a {
		if v.url != "" {
			// put a token into the rq channel (max = 20)
			rq.tokens <- struct{}{}
			url := v.url
			categroyName := k
			go func(url string, topic string) {
				// release the token
				defer func() { <-rq.tokens }()

				// Try to process one of the url
				fmt.Println(url)
				rl, err := fetchingUrlData(url, topic)
				queue <- rl
				if err != nil {
					log.Println(err.Error())
				}
			}(url, categroyName)

		} else {
			continue
		}
	}
	go func() {
		for t := range queue {
			rq.reposList = append(rq.reposList, t...)
			log.Println(len(rq.reposList))
		}
	}()
	return
}
func (a AwesomeProjects) String() (s string) {
	//var temp =""
	for k, v := range a {
		status, ok := StatusList[v.status]
		if !ok {
			status = StatusList[0]
		}
		s += fmt.Sprintf("%s : {url:%s,fetched:%v,topic:%s}\n", k, v.url, status, v.topic)
		fmt.Printf("%+v %+v", k, v)
	}
	return s
}

func getTitleSelections(doc *goquery.Document, selections []*goquery.Selection) []*goquery.Selection {
	var titleFilter = []string{"Contents", "License"}
	doc.Find("h2").Each(func(i int, s *goquery.Selection) {
		//we need filter to remove the useless h2 title.
		if !utils.Contains(titleFilter, s.Text()) && strings.TrimSpace(s.Text()) != "" {
			selections = append(selections, s)
		}
		return
	})
	return selections
}
func getSelectionsLinks(selections []*goquery.Selection, receiver interface{}) {

	for i := 0; i < len(selections); i++ {
		topic := selections[i].Text()
		selections[i].Next().Find("ul li a").Each(func(j int, s *goquery.Selection) {
			if href, exists := s.Attr("href"); exists {
				if r, ok := receiver.(*AwesomeProjects); ok {
					r.PutTopicAndUrl(topic, s.Text(), href)
					return
				}
			}
		})
	}
}

func getSelectionsLinksToArray(selections []*goquery.Selection, rl ReposList, topic string) {

	for i := 0; i < len(selections); i++ {
		topic := selections[i].Text()
		selections[i].Next().Find("ul li a").Each(func(j int, s *goquery.Selection) {
			if href, exists := s.Attr("href"); exists {
				trimedString := strings.TrimPrefix(href, "https://github.com/")
				if trimedString == href {
					return
				}
				log.Println(href, " Name:", s.Text(), " Topic:", topic)
			}
		})
	}
}

func fetchingUrlData(url string, topic string) (rplist []Repository, err error) {
	titleSelections := []*goquery.Selection{}
	log.Println("Fetching----->" + url)
	doc, err := goquery.NewDocument(url)
	if err != nil {
		return nil, err
	}
	titleSelections = getTitleSelections(doc, titleSelections)
	getSelectionsLinksToArray(titleSelections, rplist, topic)
	return
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

	// Phase 1 Download the awesome list.
	awsomeProjects = make(map[string]*Store)
	titleSelections := []*goquery.Selection{}
	doc, err := goquery.NewDocument(root_url)
	utils.CheckErrorPanic(err)
	titleSelections = getTitleSelections(doc, titleSelections)
	getSelectionsLinks(titleSelections, &awsomeProjects)
	err = awsomeProjects.SaveToDatabase(Db)
	utils.CheckErrorPanic(err)

	//	Phase 2 Scrap each repos.

	var max_jobs = 5
	ReposQuery := ReposQuery{tokens: make(chan struct{}, max_jobs), reposList: []Repository{}}
	awsomeProjects.GetRepos(Db, &ReposQuery)

}
