package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"time"

	_ "github.com/lib/pq"
)

type User struct {
	Id       int
	Name     string
	Password string
}

type Product struct {
	Id         int    `json:"id"`
	Name       string `json:"name"`
	Price      int    `json:"price"`
	ShortDesc  string `json:"shortdesc"`
	LongDesc   string `json:"longdesc"`
	CategoryId int    `json:"categoryid"`
	Image      string `json:"image"`
}

type Category struct {
	Id    int
	Name  string
	Image string
}

var cache = map[string]User{}
var produts = []Product{}
var categories = []Category{}

func Login(page http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("html_files/login.html")
	if err != nil {
		panic(err)
	}
	tmpl.ExecuteTemplate(page, "login", nil)
}

func LoginCheck(page http.ResponseWriter, r *http.Request) {
	login := r.FormValue("login")
	password := r.FormValue("password")

	if login == "" || password == "" {
		tmpl, err := template.ParseFiles("html_files/login.html")
		if err != nil {
			panic(err)
		}
		tmpl.ExecuteTemplate(page, "login", "Имя или пароль не может быть пустым")
		return
	}

	connStr := "user=postgres password=123456 dbname=netshopgolang sslmode=disable"
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		panic(err)
	}

	hash := md5.Sum([]byte(password))
	hashedPass := hex.EncodeToString(hash[:])

	defer db.Close()
	res := db.QueryRow("SELECT * FROM public.users WHERE name = $1 AND password = $2", login, hashedPass)
	user := User{}
	err = res.Scan(&user.Id, &user.Name, &user.Password)
	if err != nil {
		tmpl, err := template.ParseFiles("html_files/login.html")
		if err != nil {
			panic(err)
		}
		tmpl.ExecuteTemplate(page, "login", "неверный логин или пароль")
		return

	}

	if user.Name == "admin" && user.Password == hashedPass {

		token := login
		hashToken := md5.Sum([]byte(token))
		hashedToken := hex.EncodeToString(hashToken[:])
		cache[hashedToken] = user
		livingTime := 120 * time.Hour
		expiration := time.Now().Add(livingTime)

		cookie := http.Cookie{Name: "token", Value: url.QueryEscape(hashedToken), Expires: expiration}
		http.SetCookie(page, &cookie)
		http.Redirect(page, r, "/products", http.StatusSeeOther)

	}
}

func Categories(page http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("html_files/category.html", "html_files/zagolovok.html")
	if err != nil {
		panic(err)
	}

	connStr := "user=postgres password=123456 dbname=netshopgolang sslmode=disable"
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		panic(err)
	}

	defer db.Close()

	res, err := db.Query("SELECT * FROM public.categories")

	if err != nil {
		panic(err)
	}

	categories = []Category{}
	for res.Next() {
		var cat Category
		err = res.Scan(&cat.Id, &cat.Name, &cat.Image)
		if err != nil {
			panic(err)
		}
		categories = append(categories, cat)
	}

	if len(cache) > 0 {
		tmpl.ExecuteTemplate(page, "category", categories)
	} else {
		http.Redirect(page, r, "/", http.StatusSeeOther)
	}

}

func Products(page http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("html_files/products.html", "html_files/zagolovok.html")
	if err != nil {
		panic(err)
	}

	connStr := "user=postgres password=123456 dbname=netshopgolang sslmode=disable"
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		panic(err)
	}

	defer db.Close()

	res, err := db.Query("SELECT * FROM public.products")

	if err != nil {
		panic(err)
	}

	produts = []Product{}
	for res.Next() {
		var prd Product
		err = res.Scan(&prd.Id, &prd.Name, &prd.Price, &prd.ShortDesc, &prd.LongDesc, &prd.CategoryId, &prd.Image)
		if err != nil {
			panic(err)
		}

		produts = append(produts, prd)

	}

	defer res.Close()

	if len(cache) > 0 {
		tmpl.ExecuteTemplate(page, "products", produts)
	} else {
		http.Redirect(page, r, "/", http.StatusSeeOther)
	}

}

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	http.HandleFunc("/", Login)
	http.HandleFunc("/login_check", LoginCheck)

	http.HandleFunc("/products", Products)
	http.HandleFunc("/category", Categories)
	http.ListenAndServe(":8081", nil)

	fmt.Print(len(cache))

}
