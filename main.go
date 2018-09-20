package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/jinzhu/gorm/dialects/mssql"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var db *gorm.DB

type MyHandler struct{}

type BaseFields struct {
	ID        uint       `json:"id" gorm:"primary_key"`
	CreatedAt time.Time  `json:"-"`
	UpdatedAt time.Time  `json:"-"`
	DeletedAt *time.Time `json:"-" sql:"index"`
}

type Product struct {
	BaseFields
	Code   string  `json:"code"`
	Colors []Color `json:"color" gorm:"foreignkey:ProductID"`
}

type Color struct {
	BaseFields
	Name      string `json:"name"`
	ProductID uint   `json:"-"`
}

type version struct {
	Version string `json:"version"`
}

type ShoeServer interface {
	status(resp http.ResponseWriter, r *http.Request)
	version(resp http.ResponseWriter, r *http.Request)
	add(resp http.ResponseWriter, r *http.Request)
	get(resp http.ResponseWriter, r *http.Request)
	delete(resp http.ResponseWriter, r *http.Request)
}

func NewHandler() MyHandler {
	return MyHandler{}
}

func handlerWrapper(h http.HandlerFunc) http.HandlerFunc {
	return basicAuth(func(resp http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer func() {
			log.Printf("Elapsed: %v", time.Since(start))
		}()

		h(resp, r)
	})
}

func (m MyHandler) ServeHTTP(resp http.ResponseWriter, r *http.Request) {
	log.Printf("URI: %s", r.RequestURI)
}

func (m *MyHandler) status(resp http.ResponseWriter, r *http.Request) {
	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(http.StatusOK)
}

func (m *MyHandler) version(resp http.ResponseWriter, r *http.Request) {
	file, err := os.Open("./VERSION")
	if err != nil {
		log.Println(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var v = ""

	for scanner.Scan() {
		fmt.Println(scanner.Text())
		v = scanner.Text()
	}

	versionJson, err := json.Marshal(version{Version: v})

	if err != nil {
		resp.Header().Set("Content-Type", "application/json")
		resp.WriteHeader(http.StatusInternalServerError)

		log.Println(err)
		return
	}

	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(http.StatusOK)
	resp.Write(versionJson)
}

func (m *MyHandler) add(resp http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		resp.Header().Set("Content-Type", "application/json")
		resp.WriteHeader(http.StatusForbidden)

		log.Println("Only the POST method is supported.")
		return
	}

	var product Product

	err := json.NewDecoder(r.Body).Decode(&product)

	if err != nil {
		resp.Header().Set("Content-Type", "application/json")
		resp.WriteHeader(http.StatusInternalServerError)

		log.Println(err)
		return
	}

	log.Println(product.Code)

	db.Create(&product)

	productJson, err := json.Marshal(product)

	if err != nil {
		resp.Header().Set("Content-Type", "application/json")
		resp.WriteHeader(http.StatusInternalServerError)

		log.Println(err)
		return
	}

	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(http.StatusCreated)
	resp.Write(productJson)
}

func (m *MyHandler) get(resp http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		resp.Header().Set("Content-Type", "application/json")
		resp.WriteHeader(http.StatusForbidden)

		log.Println("Only the POST method is supported.")
		return
	}

	var list []Product
	id := r.URL.Query().Get("id")

	if id != "" {
		db.Preload("Colors").First(&list, id)
	} else {
		db.Preload("Colors").Find(&list)
	}

	productJson, err := json.Marshal(list)

	if err != nil {
		panic(err)
	}

	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(http.StatusOK)
	resp.Write(productJson)
}

func (m MyHandler) delete(resp http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		resp.Header().Set("Content-Type", "application/json")
		resp.WriteHeader(http.StatusForbidden)

		log.Println("Only the POST method is supported.")
		return
	}

	var product Product

	err := json.NewDecoder(r.Body).Decode(&product)

	if err != nil {
		resp.Header().Set("Content-Type", "application/json")
		resp.WriteHeader(http.StatusInternalServerError)

		log.Println(err)
		return
	}

	db.Delete(&product, "id = ?", product.ID)

	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(http.StatusOK)
}

func NewServer(ss ShoeServer) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/status", handlerWrapper(ss.status))
	mux.HandleFunc("/version", handlerWrapper(ss.version))
	mux.HandleFunc("/add", handlerWrapper(ss.add))
	mux.HandleFunc("/list", handlerWrapper(ss.get))
	mux.HandleFunc("/delete", handlerWrapper(ss.delete))

	mux.Handle("/", http.FileServer(http.Dir(".")))

	return mux
}

func main() {
	var err error
	//db, err = gorm.Open("mysql", "root:mysql@tcp(127.0.0.1:3306)/shoeshop?charset=utf8&parseTime=True&loc=Local")
	db, err = gorm.Open(os.Getenv("DB_TYPE"), os.Getenv("DB_CONNECT"))

	if err != nil {
		panic(err)
	}

	defer func() {
		db.Close()
	}()
	db.LogMode(true)

	fmt.Println("Server started")
	h := NewHandler()
	mux := NewServer(&h)

	if !db.HasTable(&Product{}) {
		db.CreateTable(&Product{})
	}

	if !db.HasTable(&Color{}) {
		db.CreateTable(&Color{})
	}

	//err = http.ListenAndServe(":8080", mux)
	err = http.ListenAndServe(os.Getenv("LISTEN_PORT"), mux)

	if err != nil {
		panic(err)
	}
}
