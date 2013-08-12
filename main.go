/*
	Blogging with go, markdown and redis.
	Copyright (c) 2012 Jacob Dearing
*/
package main

import (
	"flag"
	store "github.com/dearing/blog/storage/redis"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

var conf = flag.String("conf", "blog.conf", "JSON configuration")
var gen = flag.Bool("gen", false, "generate a new config as conf is set")
var config Config

//  MAIN
func main() {

	flag.Parse()

	if *gen {
		config.GenerateConfig(*conf)
		log.Println("generated new config at", *conf)
		return
	}

	config.LoadConfig(*conf)

	if config.Verbose {
		log.Println("configuration loaded from " + *conf)
	}

	// Initialize contact with the server using our arguments or defaults.
	// TODO: failure checks error handling etc...
	store.Connect(config.RedisHost, config.RedisPass, config.RedisDB)
	store.LoadDirectory(config.ContentFolder, config.Suffix)

	//	Setup our handlers and get cracking...
	r := mux.NewRouter()
	r.HandleFunc("/", indexHandler)             // index
	r.HandleFunc("/toc", tocHandler)            // table of contents
	r.HandleFunc("/p/{id}", contentHandler)     // display a post with title
	r.HandleFunc("/e/{id}", editContentHandler) // edit a post
	r.HandleFunc("/s/{id}", saveContentHandler) // save a post
	r.HandleFunc("/login", loginHandler)        // fire up Outh2
	r.HandleFunc("/logout", logoutHander)       // ''
	r.HandleFunc("/callback", callbackHandler)  // Outh2 callback addy
	r.HandleFunc("/secret", secretPageHandler)  // simple login testing handler

	if config.EnableWWW {
		r.PathPrefix("/").Handler(http.FileServer(http.Dir(config.WWWRoot)))
		log.Println("handling static content from", config.WWWRoot)
	}
	http.Handle("/", r)

	log.Println("listening on", config.WWWHost)
	if err := http.ListenAndServe(config.WWWHost, nil); err != nil {
		log.Printf("%v\n", err)
	}
}
