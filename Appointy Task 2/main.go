package main

import (
	"net/http"
	"time"
	"encoding/json"
	"sync"
	"io/ioutil"
	"fmt"
	"strings"
)

type Article struct {
	Id 			string 		`json: "id"`
	Title 		string 		`json: "title"`
	SubTitle 	string 		`json: "subtitle"`
	Content 	string 		`json: "content"`
	Creation 	time.Time 	`json: "creation"`
}

type articlesHandler struct {
	sync.Mutex
	store map[string]Article
}

func (h *articlesHandler) articles(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		h.get(w, r)
		return
	case "POST":
		h.post(w, r)
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("method not allowed"))
		return
	}
}

func (h *articlesHandler) get(w http.ResponseWriter, r *http.Request) {
	articles := make([]Article, len(h.store))

	h.Lock()
	i := 0
	for _, article := range h.store {
		articles[i] = article
		i++
	}
	h.Unlock()

	jsonBytes, err := json.Marshal(articles)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (h *articlesHandler) getArticle(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.String(), "/")
	if len(parts) != 3 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	h.Lock()
	article, ok := h.store[parts[2]]
	h.Unlock()
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	jsonBytes, err := json.Marshal(article)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (h *articlesHandler) post(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	ct := r.Header.Get("content-type")
	if ct != "application/json" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		w.Write([]byte(fmt.Sprintf("need content-type 'application/json', but got '%s'", ct)))
		return
	}

	var article Article
	err = json.Unmarshal(bodyBytes, &article)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	article.Id = fmt.Sprintf("%d", time.Now().UnixNano())
	h.Lock()
	h.store[article.Id] = article
	defer h.Unlock()
}

func (h *articlesHandler) searchArticle(w http.ResponseWriter, r *http.Request){
	v := r.URL.Query().Get("q")

	var queries[]Article
	var titleArr[]string
	var subTitleArr[]string
	var contentArr[]string
	

	for _, item:= range h.store{
		// Title
		titleArr = strings.Split(item.Title, " ")
		for _, titleItem:= range titleArr{
			if(titleItem==v){
				queries=append(queries, item)
				break
			}
		}
		//Sub-Title
		subTitleArr = strings.Split(item.SubTitle, " ")
		for _, subtitleItem:= range subTitleArr{
			if(subtitleItem==v){
				queries=append(queries, item)
				break
			}
		}
		//Content
		contentArr = strings.Split(item.Content, " ")
		for _, contentItem:= range contentArr{
			if(contentItem==v){
				queries=append(queries, item)
				break
			}
		}
	}

	result, err:=json.Marshal(queries)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(result)
}

func newArticlesHandler() *articlesHandler {
	return &articlesHandler{
		store: map[string]Article{
			"id1": Article{
				Id:			"id1",
				Title:		"Hola!!",
				SubTitle:	"Hello!!",
				Content:	"Hola Amigos!!",
				Creation:	time.Now(),
			},
		},
	}
}

func main() {
	articlesHandler := newArticlesHandler()
	http.HandleFunc("/articles/search", articlesHandler.searchArticle)
	http.HandleFunc("/articles", articlesHandler.articles)
	http.HandleFunc("/articles/", articlesHandler.getArticle)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}