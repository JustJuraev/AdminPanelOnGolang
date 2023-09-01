package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"io"
	"path/filepath"

	"html/template"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gorilla/mux"
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

type Order struct {
	Id       int
	Address  string
	Delivery bool
	Number   string
	CartNum  string
	Time     time.Time
	UserId   int
}

type OrderItem struct {
	Id           int
	ProductId    int
	ProductName  string
	ProductCount int
	OrderId      int
}

var cache = map[string]User{}
var produts = []Product{}
var categories = []Category{}
var users = []User{}
var orders = []Order{}
var orderitem = []OrderItem{}

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

func Users(page http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("html_files/user.html", "html_files/zagolovok.html")
	if err != nil {
		panic(err)
	}

	connStr := "user=postgres password=123456 dbname=netshopgolang sslmode=disable"
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		panic(err)
	}

	defer db.Close()

	res, err := db.Query("SELECT * FROM public.users")

	if err != nil {
		panic(err)
	}

	users = []User{}
	for res.Next() {
		var user User
		err = res.Scan(&user.Id, &user.Name, &user.Password)
		if err != nil {
			panic(err)
		}
		users = append(users, user)
	}

	if len(cache) > 0 {
		tmpl.ExecuteTemplate(page, "user", users)
	} else {
		http.Redirect(page, r, "/login", http.StatusSeeOther)
	}
}

func AddProductForm(page http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("html_files/addproduct.html", "html_files/zagolovok.html")
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

	tmpl.ExecuteTemplate(page, "addproduct", categories)
}

func AddingProductPost(page http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	price := r.FormValue("price")
	shortdesc := r.FormValue("shortdesc")
	longdesc := r.FormValue("longdesc")
	categoryid := r.FormValue("categoryid")

	file, handler, err := r.FormFile("myFile")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	dst, _ := os.Create(filepath.Join("temp-images", handler.Filename))
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(page, err.Error(), http.StatusInternalServerError)
		return
	}

	connStr := "user=postgres password=123456 dbname=netshopgolang sslmode=disable"
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		panic(err)
	}

	defer db.Close()

	_, err = db.Exec("INSERT INTO public.products (name, price, shortdesc, longdesc, categoryid, image) VALUES ($1, $2, $3, $4, $5, $6)", name, price, shortdesc, longdesc, categoryid, handler.Filename)

	http.Redirect(page, r, "/products", http.StatusSeeOther)
}

func addCategoryForm(page http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("html_files/addcategory.html", "html_files/zagolovok.html")
	if err != nil {
		panic(err)
	}
	tmpl.ExecuteTemplate(page, "addcategory", nil)
}

func addCategoryPost(page http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")

	file, handler, err := r.FormFile("myFile")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	dst, _ := os.Create(filepath.Join("temp-images", handler.Filename))
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(page, err.Error(), http.StatusInternalServerError)
		return
	}

	connStr := "user=postgres password=123456 dbname=netshopgolang sslmode=disable"
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		panic(err)
	}

	defer db.Close()

	_, err = db.Exec("INSERT INTO public.categories (name, image) VALUES ($1, $2)", name, handler.Filename)

	http.Redirect(page, r, "/category", http.StatusSeeOther)
}

func GetOrderItem(page http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	tmpl, err := template.ParseFiles("html_files/getorderitem.html", "html_files/zagolovok.html")
	connStr := "user=postgres password=123456 dbname=netshopgolang sslmode=disable"
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		panic(err)
	}

	defer db.Close()

	row, err := db.Query("SELECT * FROM public.orderitems WHERE orderid = $1", id)
	orderitem := []OrderItem{}

	for row.Next() {
		var orditem OrderItem
		err = row.Scan(&orditem.Id, &orditem.ProductId, &orditem.ProductName, &orditem.ProductCount, &orditem.OrderId)
		if err != nil {
			panic(err)
		}

		orderitem = append(orderitem, orditem)
	}

	if len(cache) > 0 {
		tmpl.ExecuteTemplate(page, "getorderitems", orderitem)
	} else {
		http.Redirect(page, r, "/login", http.StatusSeeOther)
	}
}

func Orders(page http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("html_files/order.html", "html_files/zagolovok.html")
	if err != nil {
		panic(err)
	}

	connStr := "user=postgres password=123456 dbname=netshopgolang sslmode=disable"
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		panic(err)
	}

	defer db.Close()

	res, err := db.Query("SELECT * FROM public.orders")

	if err != nil {
		panic(err)
	}

	orders = []Order{}
	for res.Next() {
		var ord Order
		err = res.Scan(&ord.Id, &ord.Address, &ord.Delivery, &ord.Number, &ord.CartNum, &ord.Time, &ord.UserId)
		if err != nil {
			panic(err)
		}
		orders = append(orders, ord)
	}

	if len(cache) > 0 {
		tmpl.ExecuteTemplate(page, "order", orders)
	} else {
		http.Redirect(page, r, "/login", http.StatusSeeOther)
	}
}

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	router := mux.NewRouter()
	http.Handle("/", router)
	router.HandleFunc("/", Login)
	router.HandleFunc("/orderitem/{id:[0-9]+}", GetOrderItem)
	router.HandleFunc("/login_check", LoginCheck)
	router.HandleFunc("/products", Products)
	router.HandleFunc("/category", Categories)
	router.HandleFunc("/user", Users)
	router.HandleFunc("/order", Orders)
	router.HandleFunc("/addproduct", AddProductForm)
	router.HandleFunc("/adding_product", AddingProductPost)
	router.HandleFunc("/addcategory", addCategoryForm)
	router.HandleFunc("/adding_category", addCategoryPost)
	http.ListenAndServe(":8081", nil)

	//	fmt.Print(len(cache))

}
