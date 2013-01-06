package redis

import (
	"errors"
	"fmt"
	"github.com/russross/blackfriday"
	"github.com/vmihailenco/redis"
	"html/template"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"
)

type Post struct {
	ID       string
	Title    string
	Content  template.HTML
	Created  time.Time
	Modified time.Time
	Accessed string
}

var client *redis.Client

// TODO: error handling on failed connection
func Connect(host string, pass string, db int64) (e error) {
	client = redis.NewTCPClient(host, pass, db)
	return e
}

func Close() (e error) {
	e = client.Close()
	return e
}

// TODO: error handling
func Set(p Post) (e error) {

	//log.Print(p)
	key := fmt.Sprintf("post:" + p.ID)

	client.HSet(key, "title", p.Title)
	client.HSet(key, "content", string(p.Content))
	client.HSet(key, "created", fmt.Sprint(p.Created.Unix()))
	client.HSet(key, "modified", fmt.Sprint(p.Modified.Unix()))
	client.HIncrBy(key, "accessed", 1)

	return e

}

func Get(id string, incr bool) (p Post, e error) {

	key := fmt.Sprintf("post:%s", id)

	if !client.Exists(key).Val() {
		return p, errors.New("key doesn't exist : " + key)
	}

	get := client.HGetAll(key)
	e = get.Err()
	if e != nil {
		return p, e
	}

	v := get.Val()

	// Would think that there could be a mapping here in the github.com/vmihailenco/redis library?
	con := map[string]string{}
	for i := 0; i < len(v); i += 2 {
		con[v[i]] = v[i+1]
	}

	created, e := strconv.ParseInt(con["created"], 10, 64)
	if e != nil {
		return p, e
	}

	mod, e := strconv.ParseInt(con["created"], 10, 64)
	if e != nil {
		return p, e
	}

	p.ID = id
	p.Title = con["title"]
	p.Content = template.HTML(con["content"])
	p.Created = time.Unix(created, 0)
	p.Modified = time.Unix(mod, 0)
	p.Accessed = con["accessed"]

	if incr {
		client.HIncrBy(key, "accessed", 1)
	}

	return p, e
}

func Del(id string) (e error) {
	key := fmt.Sprintf("post:%s", id)
	client.Del(key)
	return e
}

func Keys(pattern string) (keys *redis.StringSliceReq) {
	return client.Keys(pattern)
}

func getHTML(content string) template.HTML {
	return template.HTML(blackfriday.MarkdownCommon([]byte(content)))
}

func LoadDirectory(path string) (e error) {
	x, e := ioutil.ReadDir(path)

	if e != nil {
		return e
	}

	for _, z := range x {
		if !z.IsDir() {

			b, e := ioutil.ReadFile(path + z.Name())
			if e != nil {
				log.Println(e)
			}

			//id := client.Incr("global:nextPostID")
			p := Post{
				//ID:       fmt.Sprintf("%v", id.Val()),
				ID:       z.Name(),
				Title:    strings.TrimRight(z.Name(), ".md"),
				Content:  getHTML(string(b)),
				Created:  time.Now(),
				Modified: z.ModTime(),
				Accessed: "0",
			}
			Set(p)
		}
	}

	return e
}
