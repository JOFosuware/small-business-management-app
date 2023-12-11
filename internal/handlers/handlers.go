package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/jofosuware/small-business-management-app/internal/config"
	"github.com/jofosuware/small-business-management-app/internal/driver"
	"github.com/jofosuware/small-business-management-app/internal/forms"
	"github.com/jofosuware/small-business-management-app/internal/helpers"
	"github.com/jofosuware/small-business-management-app/internal/models"
	"github.com/jofosuware/small-business-management-app/internal/render"
	"github.com/jofosuware/small-business-management-app/internal/repository"
	"github.com/jofosuware/small-business-management-app/internal/repository/dbrepo"
	"golang.org/x/crypto/bcrypt"
)

var Repo *Repository

type Repository struct {
	App *config.AppConfig
	DB  repository.DatabaseRepo
}

// NewRepo creates a new repository
func NewRepo(a *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewPostgresRepo(db.SQL, a),
	}
}

// NewHandlers set the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

// Home is the home page handler
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "login.page.html", &models.TemplateData{
		Form: forms.New(nil),
	})

}

// Dashboard is the dashboard handler
func (m *Repository) Dashboard(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]interface{})
	metaData := models.FormMetaData{
		Url: r.URL.Path,
	}

	data["metadata"] = metaData
	render.Template(w, r, "dashboard.page.html", &models.TemplateData{
		Data: data,
	})
}

// PostLogin handles logging the user in
func (m *Repository) PostLogin(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}

	username := r.Form.Get("username")
	password := r.Form.Get("password")

	form := forms.New(r.PostForm)
	form.Required("username", "password")

	if !form.Valid() {
		render.Template(w, r, "login.page.html", &models.TemplateData{
			Form: form,
		})
		fmt.Println("Forms error", form)
		return
	}

	user, err := m.DB.Authenticate(username, password)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Invalid login credentials")
		m.App.ErrorLog.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if password == "OseePassword" {
		m.App.Session.Put(r.Context(), "user", user)
		render.Template(w, r, "reset.page.html", &models.TemplateData{})
		m.App.Session.Put(r.Context(), "user_id", user.ID)
		return
	}

	m.App.Session.Put(r.Context(), "user_id", user.ID)
	m.App.Session.Put(r.Context(), "user", user)
	m.App.Session.Put(r.Context(), "flash", "Logged in successfully")
	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
}

// User
// UserForm handles request for user form
func (m *Repository) UserForm(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]interface{})

	if r.URL.Path == "/admin/edit-user" {
		user, _ := m.App.Session.Get(r.Context(), "user").(models.User)
		metaData := models.FormMetaData{
			Section: "User",
			Message: "Edit User Information",
			Button:  "Update User",
			Url:     "/admin/edit-user",
		}

		data["usr"] = user
		data["metadata"] = metaData
	} else {
		metaData := models.FormMetaData{
			Section: "User",
			Message: "Add User Information",
			Button:  "Create User",
			Url:     "/admin/signup",
		}
		data["usr"] = models.User{}
		data["metadata"] = metaData
	}

	render.Template(w, r, "userform.page.html", &models.TemplateData{
		Data: data,
		Form: forms.New(nil),
	})
}

// PostUser ceates user
func (m *Repository) PostUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't processing form!")
		http.Redirect(w, r, "/admin/signup", http.StatusSeeOther)
		m.App.ErrorLog.Println("Form Parse error", err)
		return
	}

	firstName := r.Form.Get("firstname")
	lastName := r.Form.Get("lastname")
	username := r.Form.Get("username")
	userImage, _, _ := r.FormFile("userPhoto")
	defer userImage.Close()
	password := "OseePassword"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Couldn't process password! try again")
		http.Redirect(w, r, "/admin/signup", http.StatusSeeOther)
		m.App.ErrorLog.Println("error hashing password", err)
		return
	}

	usrImage, err := helpers.ProcessImage(userImage)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't process custumer photo")
		http.Redirect(w, r, "/admin/signup", http.StatusSeeOther)
		m.App.ErrorLog.Println(err)
		return
	}

	user := models.User{
		FirstName:   firstName,
		LastName:    lastName,
		Username:    username,
		Password:    string(hashedPassword),
		AccessLevel: r.Form.Get("accesslevel"),
		Image:       usrImage,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	form := forms.New(r.PostForm)
	form.Required("firstname", "lastname", "username")

	if !form.Valid() {
		m.App.Session.Put(r.Context(), "error", "All fields need to be filled!")
		http.Redirect(w, r, "/admin/signup", http.StatusSeeOther)
		m.App.ErrorLog.Println("empty form field", err)
		return
	}

	u, err := m.DB.FetchUser(username)

	if err != nil && err.Error() != "sql: no rows in result set" {
		m.App.Session.Put(r.Context(), "error", "Internal server error! try again")
		http.Redirect(w, r, "/admin/signup", http.StatusSeeOther)
		m.App.ErrorLog.Println(err)
		return
	}

	if u.Username == username {
		m.App.Session.Put(r.Context(), "error", "Username already exist!, choose another one")
		http.Redirect(w, r, "/admin/signup", http.StatusSeeOther)
		return
	}

	_, err = m.DB.InsertUser(user)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "User couldn't be saved! try again")
		http.Redirect(w, r, "/admin/signup", http.StatusSeeOther)
		m.App.ErrorLog.Println(err)
		return
	}

	m.App.Session.Put(r.Context(), "flash", fmt.Sprintf("User %s created!", u.Username))
	http.Redirect(w, r, "/admin/signup", http.StatusSeeOther)
}

// PostReset handles request for reseting user password
func (m *Repository) PostReset(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]any)
	data["error"] = ""
	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't processing form!")
		render.Template(w, r, "reset.page.html", &models.TemplateData{
			Data: data,
		})
		m.App.ErrorLog.Println("Form Parse error", err)
		return
	}

	user, ok := m.App.Session.Get(r.Context(), "user").(models.User)
	if !ok {
		m.App.Session.Put(r.Context(), "error", "can't find you in the session, reload the app!")
		render.Template(w, r, "reset.page.html", &models.TemplateData{
			Data: data,
		})
		m.App.ErrorLog.Println("User session error: ", err)
		return
	}

	pwd := r.Form.Get("password")
	rPwd := r.Form.Get("repeatPassword")

	if pwd != rPwd {
		data["error"] = "Passwords do not match, try again!"
		render.Template(w, r, "reset.page.html", &models.TemplateData{
			Data: data,
		})
		//m.App.Session.Put(r.Context(), "error", "passwords do not match, try again!")
		m.App.ErrorLog.Println("Password mismatch")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Couldn't process password! try again")
		render.Template(w, r, "reset.page.html", &models.TemplateData{
			Data: data,
		})
		m.App.ErrorLog.Println("error hashing password", err)
		return
	}

	usr := models.User{
		Username: user.Username,
		Password: string(hashedPassword),
	}

	form := forms.New(r.PostForm)
	form.Required("password", "repeatPassword")

	if !form.Valid() {
		m.App.Session.Put(r.Context(), "error", "All fields need to be filled!")
		render.Template(w, r, "reset.page.html", &models.TemplateData{
			Data: data,
		})
		m.App.ErrorLog.Println("empty form field")
		return
	}

	err = m.DB.ResetUser(usr)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Password reset unsuccessful, try again!")
		render.Template(w, r, "reset.page.html", &models.TemplateData{
			Data: data,
		})
		m.App.ErrorLog.Println("Database error: ", err)
		return
	}

	m.App.Session.Put(r.Context(), "flash", "Logged in successfully")
	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
}

// PostDeveloper ceates one superuser for the developer
func (m *Repository) PostDeveloper(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't processing form!")
		http.Redirect(w, r, "/signup", http.StatusSeeOther)
		m.App.ErrorLog.Println("Form Parse error", err)
		return
	}

	firstName := r.Form.Get("firstname")
	lastName := r.Form.Get("lastname")
	username := r.Form.Get("username")
	password := r.Form.Get("password")
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Couldn't process password! try again")
		http.Redirect(w, r, "/signup", http.StatusSeeOther)
		m.App.ErrorLog.Println("error hashing password", err)
		return
	}

	user := models.User{
		FirstName:   firstName,
		LastName:    lastName,
		Username:    username,
		Password:    string(hashedPassword),
		AccessLevel: r.Form.Get("accesslevel"),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	form := forms.New(r.PostForm)
	form.Required("firstname", "lastname", "username", "password")

	if !form.Valid() {
		m.App.Session.Put(r.Context(), "error", "All fields need to be filled!")
		http.Redirect(w, r, "/signup", http.StatusSeeOther)
		m.App.ErrorLog.Println("empty form field", err)
		return
	}

	u, err := m.DB.FetchUser(username)

	if err != nil && err.Error() != "sql: no rows in result set" {
		m.App.Session.Put(r.Context(), "error", "Internal server error! try again")
		http.Redirect(w, r, "/signup", http.StatusSeeOther)
		m.App.ErrorLog.Println(err)
		return
	}

	if u.Username == username {
		m.App.Session.Put(r.Context(), "error", "Username already exist!, choose another one")
		http.Redirect(w, r, "/signup", http.StatusSeeOther)
		return
	}

	_, err = m.DB.InsertUser(user)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "User couldn't be saved! try again")
		http.Redirect(w, r, "/signup", http.StatusSeeOther)
		m.App.ErrorLog.Println(err)
		return
	}

	m.App.Session.Put(r.Context(), "flash", fmt.Sprintf("User %s created!", u.Username))
	http.Redirect(w, r, "/signup", http.StatusSeeOther)
}

// Logout logs a user out
func (m *Repository) Logout(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.Destroy(r.Context())
	_ = m.App.Session.RenewToken(r.Context())

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Products information processing
// ProductForm is a handler that returns product creation form
func (m *Repository) ProductForm(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]interface{})

	if r.URL.Path == "/admin/edit-product" {
		prod, _ := m.App.Session.Pop(r.Context(), "product").(models.Product)

		metaData := models.FormMetaData{
			Message: "Edit Product Information",
			Button:  "Patch Product",
			Url:     "/admin/edit-product",
			Section: "Product",
		}
		data["product"] = prod
		data["metadata"] = metaData
	} else {
		metaData := models.FormMetaData{
			Message: "Enter Product Information",
			Button:  "Post Product",
			Url:     "/admin/add-product",
			Section: "Product",
		}
		data["product"] = models.Product{}
		data["metadata"] = metaData
	}

	render.Template(w, r, "addproduct.page.html", &models.TemplateData{
		Form: forms.New(nil),
		Data: data,
	})
}

// AddProduct is handler for handling product creation requests
func (m *Repository) AddProduct(w http.ResponseWriter, r *http.Request) {
	userId, ok := m.App.Session.Get(r.Context(), "user_id").(int)

	if !ok {
		m.App.Session.Put(r.Context(), "error", "internal server error")
		http.Redirect(w, r, "/admin/add-product", http.StatusSeeOther)
		m.App.ErrorLog.Println("Failed casting userid to int from session")
		return
	}

	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't processing form!")
		http.Redirect(w, r, "/admin/add-product", http.StatusSeeOther)
		m.App.ErrorLog.Println("Form Parse error", err)
		return
	}

	price, err := strconv.ParseFloat(r.Form.Get("price"), 64)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "invalid data!")
		http.Redirect(w, r, "/admin/add-product", http.StatusSeeOther)
		m.App.ErrorLog.Println("Price Parse error", err)
		return
	}

	stock, err := strconv.Atoi(r.Form.Get("stock"))
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "invalid data!")
		http.Redirect(w, r, "/admin/add-product", http.StatusSeeOther)
		m.App.ErrorLog.Println("Stock Parse error", err)
		return
	}

	product := models.Product{
		Serial:      r.Form.Get("serial"),
		Name:        r.Form.Get("name"),
		Description: r.Form.Get("description"),
		Price:       price,
		Units:       int32(stock),
		UserId:      userId,
	}

	form := forms.New(r.PostForm)
	data := make(map[string]interface{})

	form.Required("name", "description", "price", "stock")
	if !form.Valid() {
		data["product"] = product

		render.Template(w, r, "addproduct.page.html", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	p, _ := m.DB.FetchProduct(product.Serial)
	if p.Serial == product.Serial {
		m.App.Session.Put(r.Context(), "error", "Product with same serial number already exists")
		http.Redirect(w, r, "/admin/add-product", http.StatusSeeOther)
		m.App.ErrorLog.Println("Product with same serial number already exists")
		return
	}

	product, err = m.DB.InsertProduct(product)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Product could not be inserted")
		http.Redirect(w, r, "/admin/add-product", http.StatusSeeOther)
		m.App.ErrorLog.Println(err)
		return
	}

	product.Price = price

	data["product"] = product

	m.App.Session.Put(r.Context(), "flash", "Product inserted!")
	m.App.Session.Put(r.Context(), "product", product)
	render.Template(w, r, "displayproduct.page.html", &models.TemplateData{
		Data: data,
	})
}

// UpdateProduct update product with changes by ID
func (m *Repository) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't processing form!")
		http.Redirect(w, r, "/admin/add-product", http.StatusSeeOther)
		m.App.ErrorLog.Println("Error parsing form")
		return
	}

	price, err := strconv.ParseFloat(r.Form.Get("price"), 64)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "invalid data!")
		http.Redirect(w, r, "/admin/add-product", http.StatusSeeOther)
		m.App.ErrorLog.Println("Error parsing int for price")
		return
	}

	stock, err := strconv.Atoi(r.Form.Get("stock"))
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "invalid data!")
		http.Redirect(w, r, "/admin/add-product", http.StatusSeeOther)
		m.App.ErrorLog.Println("Error parsing int for stock")
		return
	}

	prod_id, err := strconv.Atoi(r.Form.Get("prod_id"))
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "invalid data!")
		http.Redirect(w, r, "/admin/add-product", http.StatusSeeOther)
		m.App.ErrorLog.Println("Error parsing int for product ID")
		return
	}

	userId := m.App.Session.Get(r.Context(), "user_id").(int)
	product := models.Product{
		ID:          prod_id,
		Serial:      r.Form.Get("serial"),
		Name:        r.Form.Get("name"),
		Description: r.Form.Get("description"),
		Price:       price,
		Units:       int32(stock),
		UserId:      userId,
	}

	form := forms.New(r.PostForm)
	data := make(map[string]interface{})

	form.Required("prod_id", "name", "description", "price", "stock")
	if !form.Valid() {
		data["product"] = product

		render.Template(w, r, "addproduct.page.html", &models.TemplateData{
			Form: form,
			Data: data,
		})
		m.App.ErrorLog.Println("Invalid form values")
		return
	}

	err = m.DB.UpdateProduct(product)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Product could not be updated")
		http.Redirect(w, r, "/admin/add-product", http.StatusSeeOther)
		m.App.ErrorLog.Println(err)
		return
	}

	data["product"] = product
	m.App.Session.Put(r.Context(), "product", product)
	m.App.Session.Put(r.Context(), "flash", "Product updated!")
	render.Template(w, r, "displayproduct.page.html", &models.TemplateData{
		Data: data,
	})
}

// SearchProduct handlers the request for Product form
func (m *Repository) SearchProduct(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]interface{})

	if r.URL.Path == "/admin/search-product" {
		metaData := models.FormMetaData{
			Url:     "/admin/edit-form",
			Section: "Product",
			Message: "Search Product",
			Button:  "Search",
		}
		data["metadata"] = metaData
	} else {
		metaData := models.FormMetaData{
			Url:     "/admin/delete-product",
			Section: "Product",
			Message: "Enter Product Serial",
			Button:  "Delete",
		}
		data["metadata"] = metaData
	}
	render.Template(w, r, "editproduct.page.html", &models.TemplateData{
		Data: data,
		Form: forms.New(nil),
	})
}

// EditProductForm handlers edit product form request
func (m *Repository) EditProductForm(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't process form")
		http.Redirect(w, r, "/admin/search-product", http.StatusSeeOther)
		fmt.Println("Error Parsing form ", err)
		return
	}

	form := forms.New(r.PostForm)
	form.Required("serial")

	if !form.Valid() {
		m.App.Session.Put(r.Context(), "error", "form field can't be empty")
		render.Template(w, r, "editproduct.page.html", &models.TemplateData{
			Form: form,
		})
		fmt.Println("Error in form values ", err)
		return
	}

	p, err := m.DB.FetchProduct(r.Form.Get("serial"))
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "No product with such serial number")
		http.Redirect(w, r, "/admin/search-product", http.StatusSeeOther)
		//fmt.Println("Error Querying database ", err)
		return
	}

	m.App.Session.Put(r.Context(), "product", p)
	http.Redirect(w, r, "/admin/edit-product", http.StatusSeeOther)
}

// ListProducts handles request for products in the database
func (m *Repository) ListProducts(w http.ResponseWriter, r *http.Request) {
	page := chi.URLParam(r, "page")
	pg, _ := strconv.Atoi(page)
	if pg == 0 {
		pg = 1
	}
	meta := models.FormMetaData{
		Section: "Product",
		Url:     "/admin/list-products/1",
	}

	data := make(map[string]any)

	prods, err := m.DB.FetchProductByPage(pg)

	var p []models.Product
	for _, v := range prods {
		v.Price = helpers.ToDecimalPlace(v.Price, 2)
		p = append(p, v)
	}
	data["products"] = p
	data["metadata"] = meta
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Products cannot be fetched!")
		m.App.ErrorLog.Println("Products cannot be fetched!")
		render.Template(w, r, "displayproducts.page.html", &models.TemplateData{
			Data: data,
		})
		return
	}

	if len(prods) == 0 {
		m.App.Session.Put(r.Context(), "error", "No product was found!")
		render.Template(w, r, "displayproducts.page.html", &models.TemplateData{
			Data: data,
		})
		return
	}

	render.Template(w, r, "displayproducts.page.html", &models.TemplateData{
		Data: data,
	})
}

// DeleteProduct handlers request for delete product form
func (m *Repository) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't process form!")
		http.Redirect(w, r, "/admin/delete-product", http.StatusSeeOther)
		return
	}

	form := forms.New(r.PostForm)
	form.Required("serial")

	if !form.Valid() {
		m.App.Session.Put(r.Context(), "error", "Invalid form data!")
		http.Redirect(w, r, "/admin/delete-product", http.StatusSeeOther)
		return
	}

	prod, err := m.DB.FetchProduct(r.Form.Get("serial"))
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "No product with such serial number found")
		http.Redirect(w, r, "/admin/delete-product", http.StatusSeeOther)
		return
	}

	if prod.Serial != r.Form.Get("serial") {
		m.App.Session.Put(r.Context(), "error", "No product with such serial number found")
		http.Redirect(w, r, "/admin/delete-product", http.StatusSeeOther)
		return
	}

	err = m.DB.DeleteProduct(r.Form.Get("serial"))
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "No product with such serial number found")
		http.Redirect(w, r, "/admin/delete-product", http.StatusSeeOther)
		return
	}

	m.App.Session.Put(r.Context(), "flash", "Product deleted successfully")
	http.Redirect(w, r, "/admin/delete-product", http.StatusSeeOther)
}

// IncreaseQtyForm serves request for form for increasing product quantity
func (m *Repository) IncreaseQtyForm(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]interface{})

	metaData := models.FormMetaData{
		Message: "Add Product Quantity",
		Button:  "Post Quantity",
		Url:     "/admin/increase-qty",
		Section: "Product",
	}

	data["metadata"] = metaData

	render.Template(w, r, "increaseQtyform.page.html", &models.TemplateData{
		Form: forms.New(nil),
		Data: data,
	})
}

// PostIncreaseQty process product quantity addition request
func (m *Repository) PostIncreaseQty(w http.ResponseWriter, r *http.Request) {
	userId, ok := m.App.Session.Get(r.Context(), "user_id").(int)
	data := make(map[string]interface{})
	if !ok {
		m.App.Session.Put(r.Context(), "error", "Failed to get user ID")
		http.Redirect(w, r, "/admin/increase-qty", http.StatusSeeOther)
		m.App.ErrorLog.Println("Failed to get user ID")
		return
	}

	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't process form")
		http.Redirect(w, r, "/admin/increase-qty", http.StatusSeeOther)
		log.Println(err)
		return
	}

	qty, _ := strconv.Atoi(r.Form.Get("quantity"))
	p := models.Product{
		Serial: r.Form.Get("serial"),
		Units:  int32(qty),
		UserId: userId,
	}

	metaData := models.FormMetaData{
		Message: "Add Product Quantity",
		Button:  "Post Quantity",
		Url:     "/admin/increase-qty",
		Section: "Product",
	}
	data["product"] = p
	data["metadata"] = metaData

	form := forms.New(r.PostForm)
	form.Required("serial", "quantity")
	if !form.Valid() {
		render.Template(w, r, "increaseQtyform.page.html", &models.TemplateData{
			Form: form,
			Data: data,
		})
	}

	prod, err := m.DB.FetchProduct(p.Serial)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Failed retrieving the product, check the serial again!")
		http.Redirect(w, r, "/admin/increase-qty", http.StatusSeeOther)
		m.App.ErrorLog.Println(err)
		return
	}

	if prod.Serial != p.Serial {
		m.App.Session.Put(r.Context(), "error", "Sorry, couldn't find a product with this serial number!")
		http.Redirect(w, r, "/admin/increase-qty", http.StatusSeeOther)
		m.App.ErrorLog.Println(err)
		return
	}

	err = m.DB.IncreaseQuantity(p)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Failed to insert customer items purchased")
		http.Redirect(w, r, "/admin/increase-qty", http.StatusSeeOther)
		m.App.ErrorLog.Println(err)
		return
	}

	m.App.Session.Put(r.Context(), "flash", fmt.Sprintf("Product with serial number %s quantity increased!", p.Serial))
	render.Template(w, r, "increaseQtyform.page.html", &models.TemplateData{
		Form: form,
		Data: data,
	})
}

// Contract Section handles Contract processing
// ContractForm handler handles contract form request
func (m *Repository) CustomerForm(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]interface{})

	if r.URL.Path == "/admin/edit-contract" {
		cust, _ := m.App.Session.Get(r.Context(), "customer").(models.Customer)
		metaData := models.FormMetaData{
			Section: "Contract",
			Message: "Edit Customer Information",
			Button:  "Update Information",
			Url:     "/admin/update-contract",
		}

		data["customer"] = cust
		data["metadata"] = metaData
	} else {
		metaData := models.FormMetaData{
			Section: "Contract",
			Message: "Add Customer Information",
			Button:  "Add Information",
			Url:     "/admin/add-contract",
		}
		data["customer"] = models.Customer{}
		data["metadata"] = metaData
	}

	render.Template(w, r, "customerform.page.html", &models.TemplateData{
		Data: data,
		Form: forms.New(nil),
	})
}

// PostCustomer handles customer creation request
func (m *Repository) PostCustomer(w http.ResponseWriter, r *http.Request) {
	userId, ok := m.App.Session.Get(r.Context(), "user_id").(int)
	if !ok {
		m.App.Session.Put(r.Context(), "error", "Failed to get user ID")
		http.Redirect(w, r, "/admin/add-contract", http.StatusSeeOther)
		m.App.ErrorLog.Println("Failed to get user ID")
		return
	}
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't process form")
		http.Redirect(w, r, "/admin/add-contract", http.StatusSeeOther)
		log.Println(err)
		return
	}

	customerId := r.Form.Get("customerId")
	idType := r.Form.Get("idtype")
	cardImage, _, _ := r.FormFile("cardPhoto")
	custImage, _, _ := r.FormFile("custPhoto")
	firstName := r.Form.Get("firstname")
	lastName := r.Form.Get("lastname")
	hAddress := r.Form.Get("houseaddress")
	phone := r.Form.Get("phone")
	location := r.Form.Get("location")
	landmark := r.Form.Get("landmark")
	months, _ := strconv.Atoi(r.Form.Get("months"))
	agreement := r.Form.Get("agreement")

	ctsImage, err := helpers.ProcessImage(custImage)
	defer custImage.Close()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't process custumer photo")
		http.Redirect(w, r, "/admin/add-contract", http.StatusSeeOther)
		m.App.ErrorLog.Println(err)
		return
	}

	crdImage, err := helpers.ProcessImage(cardImage)
	defer cardImage.Close()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't process ID card photo")
		http.Redirect(w, r, "/admin/add-contract", http.StatusSeeOther)
		m.App.ErrorLog.Println(err)
		return
	}

	// do necessary conversion
	p, _ := strconv.Atoi(phone)

	c := models.Customer{
		CustomerId:   customerId,
		CustImage:    ctsImage,
		IDType:       idType,
		CardImage:    crdImage,
		FirstName:    firstName,
		LastName:     lastName,
		Phone:        p,
		HouseAddress: hAddress,
		Location:     location,
		Landmark:     landmark,
		Status:       "on_contract",
		Months:       months,
		Agreement:    agreement,
		UserId:       userId,
	}

	form := forms.New(r.Form)
	data := make(map[string]interface{})
	form.Required("customerId", "idtype", "firstname", "lastname", "houseaddress", "phone", "location", "landmark", "agreement")

	if !form.Valid() {
		m.App.Session.Put(r.Context(), "error", "One of the fields is empty")
		metaData := models.FormMetaData{
			Section: "Contract",
			Message: "Add Customer Information",
			Button:  "Add Information",
			Url:     "/admin/add-contract",
		}
		data["customer"] = c
		data["metadata"] = metaData
		render.Template(w, r, "customerform.page.html", &models.TemplateData{
			Form: form,
			Data: data,
		})
		m.App.ErrorLog.Println("One of the fields is empty")
		return
	}

	cust, _ := m.DB.FetchCustomer(c.CustomerId)

	if cust.CustomerId == c.CustomerId {
		m.App.Session.Put(r.Context(), "error", "Customer with such ID already exist")
		http.Redirect(w, r, "/admin/add-contract", http.StatusSeeOther)
		m.App.InfoLog.Println("Customer with such ID already exist")
		return
	}

	err = m.DB.InsertCustomer(c)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Customer information could not be inserted!")
		http.Redirect(w, r, "/admin/add-contract", http.StatusSeeOther)
		m.App.ErrorLog.Println(err)
		return
	}

	data["customer"] = c
	m.App.Session.Put(r.Context(), "customer", c)
	m.App.Session.Put(r.Context(), "customerId", c.CustomerId)
	m.App.Session.Put(r.Context(), "flash", "Customer inserted successfully")
	render.Template(w, r, "displayCustomer.page.html", &models.TemplateData{
		Data: data,
	})
}

// UpdateCustomer update customer information with changes by ID
func (m *Repository) UpdateCustomer(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't processing form!")
		http.Redirect(w, r, "/admin/add-contract", http.StatusSeeOther)
		m.App.ErrorLog.Println("Error parsing form")
		return
	}

	userId := m.App.Session.Get(r.Context(), "user_id").(int)
	customerId := r.Form.Get("customerId")
	idType := r.Form.Get("idtype")
	firstName := r.Form.Get("firstname")
	lastName := r.Form.Get("lastname")
	hAddress := r.Form.Get("houseaddress")
	phone := r.Form.Get("phone")
	location := r.Form.Get("location")
	landmark := r.Form.Get("landmark")
	months, _ := strconv.Atoi(strings.TrimSpace(r.Form.Get("months")))
	agreement := r.Form.Get("agreement")
	custImage, _, _ := r.FormFile("custPhoto")
	cardImage, _, _ := r.FormFile("cardPhoto")

	// do necessary conversion
	p, _ := strconv.Atoi(strings.TrimSpace(phone))
	id, _ := strconv.Atoi(r.Form.Get("cust_id"))

	defer custImage.Close()
	defer cardImage.Close()

	ctsImg, err := helpers.ProcessImage(custImage)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "error processing customer's image!")
		http.Redirect(w, r, "/admin/add-contract", http.StatusSeeOther)
		m.App.ErrorLog.Println(err)
		return
	}

	crdImg, err := helpers.ProcessImage(cardImage)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "error processing id card's image!")
		http.Redirect(w, r, "/admin/add-contract", http.StatusSeeOther)
		m.App.ErrorLog.Println(err)
		return
	}

	c := models.Customer{
		ID:           id,
		CustomerId:   customerId,
		CustImage:    ctsImg,
		IDType:       idType,
		CardImage:    crdImg,
		FirstName:    firstName,
		LastName:     lastName,
		Phone:        p,
		HouseAddress: hAddress,
		Location:     location,
		Landmark:     landmark,
		Status:       "on_contract",
		Agreement:    agreement,
		Months:       months,
		UserId:       userId,
	}

	form := forms.New(r.Form)
	data := make(map[string]interface{})
	form.Required("customerId", "idtype", "firstname", "lastname", "houseaddress", "phone",
		"location", "landmark", "agreement")

	if !form.Valid() {
		m.App.Session.Put(r.Context(), "error", "One of the fields is empty")
		metaData := models.FormMetaData{
			Section: "Contract",
			Message: "Add Customer Information",
			Button:  "Add Information",
			Url:     "/admin/add-contract",
		}
		data["customer"] = c
		data["metadata"] = metaData
		render.Template(w, r, "customerform.page.html", &models.TemplateData{
			Form: form,
			Data: data,
		})
		m.App.ErrorLog.Println("One of the fields is empty")
		return
	}

	err = m.DB.UpdateCustomer(c)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Customer informaction could not be updated")
		http.Redirect(w, r, "/admin/add-contract", http.StatusSeeOther)
		m.App.ErrorLog.Println(err)
		return
	}

	data["customer"] = c
	m.App.Session.Put(r.Context(), "customer", c)
	m.App.Session.Put(r.Context(), "customerId", c.CustomerId)
	m.App.Session.Put(r.Context(), "flash", "Customer updated successfully")
	render.Template(w, r, "displayCustomer.page.html", &models.TemplateData{
		Data: data,
	})
}

func (m *Repository) GetWitnessForm(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]interface{})

	if r.URL.Path == "/admin/edit-witness" {
		witn, _ := m.App.Session.Pop(r.Context(), "witness").(models.Witness)
		meta := models.FormMetaData{
			Message: "Eidt Witness Information",
			Button:  "Edit Witness Information",
			Url:     "/admin/update-witness",
			Section: "Contract",
		}

		data["witness"] = witn
		data["metadata"] = meta
	} else {
		meta := models.FormMetaData{
			Message: "Add Witness Information",
			Button:  "Add Witness Information",
			Url:     "/admin/add-witness",
			Section: "Contract",
		}
		data["witness"] = models.Witness{}
		data["metadata"] = meta
	}

	render.Template(w, r, "witnessform.page.html", &models.TemplateData{
		Form: forms.New(nil),
		Data: data,
	})
}

// PostWitness handles post request for witness
func (m *Repository) PostWitness(w http.ResponseWriter, r *http.Request) {
	userId, ok := m.App.Session.Get(r.Context(), "user_id").(int)
	if !ok {
		m.App.Session.Put(r.Context(), "error", "Failed to get user ID")
		http.Redirect(w, r, "/admin/add-witness", http.StatusSeeOther)
		m.App.ErrorLog.Println("Failed to get user ID")
		return
	}

	customerId, ok := m.App.Session.Get(r.Context(), "customerId").(string)
	if !ok {
		m.App.Session.Put(r.Context(), "error", "Failed to get customer ID")
		http.Redirect(w, r, "/admin/add-witness", http.StatusSeeOther)
		m.App.ErrorLog.Println("Failed to get customer ID")
		return
	}

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't process form")
		http.Redirect(w, r, "/admin/add-witness", http.StatusSeeOther)
		m.App.ErrorLog.Println(err)
		return
	}

	p, _ := strconv.Atoi(r.Form.Get("phone"))
	witnessPhoto, _, _ := r.FormFile("witnessPhoto")

	defer witnessPhoto.Close()
	wtnImg, err := helpers.ProcessImage(witnessPhoto)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "error processing witness image")
		http.Redirect(w, r, "/admin/add-witness", http.StatusSeeOther)
		m.App.ErrorLog.Println(err)
		return
	}

	witn := models.Witness{
		CustomerId: customerId,
		FirstName:  r.Form.Get("firstname"),
		LastName:   r.Form.Get("lastname"),
		Phone:      p,
		Terms:      r.Form.Get("terms"),
		Image:      wtnImg,
		UserId:     userId,
	}

	data := make(map[string]interface{})
	form := forms.New(r.Form)
	form.Required("firstname", "lastname", "phone", "terms")
	if !form.Valid() {
		meta := models.FormMetaData{
			Message: "Add Witness Information",
			Button:  "Add Witness Information",
			Url:     "/admin/add-witness",
			Section: "Contract",
		}

		data["witness"] = witn
		data["metadata"] = meta
		render.Template(w, r, "witnessform.page.html", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	wtn, _ := m.DB.FetchWitness(witn.CustomerId)
	if wtn.CustomerId == witn.CustomerId {
		m.App.Session.Put(r.Context(), "error", "Witness with such ID already exist")
		http.Redirect(w, r, "/admin/add-witness", http.StatusSeeOther)
		m.App.InfoLog.Println("Witness with such ID already exist")
		return
	}

	wtn, err = m.DB.InsertWitnessData(witn)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Witness information could not be inserted!")
		http.Redirect(w, r, "/admin/add-witness", http.StatusSeeOther)
		m.App.ErrorLog.Println(err)
		return
	}

	data["witness"] = wtn
	m.App.Session.Put(r.Context(), "witness", wtn)
	m.App.Session.Put(r.Context(), "customerId", wtn.CustomerId)
	m.App.Session.Put(r.Context(), "flash", "Witness information inserted successfully")
	render.Template(w, r, "displayWitness.page.html", &models.TemplateData{
		Data: data,
	})
}

// UpdateWitness update witness information with changes by ID
func (m *Repository) UpdateWitness(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't processing form!")
		http.Redirect(w, r, "/admin/add-witness", http.StatusSeeOther)
		m.App.ErrorLog.Println(err)
		return
	}

	userId := m.App.Session.Get(r.Context(), "user_id").(int)
	p, _ := strconv.Atoi(r.Form.Get("phone"))
	witnessImage, _, _ := r.FormFile("witnessPhoto")
	defer witnessImage.Close()

	witnImg, err := helpers.ProcessImage(witnessImage)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "error processing witness's image!")
		http.Redirect(w, r, "/admin/add-witness", http.StatusSeeOther)
		m.App.ErrorLog.Println(err)
		return
	}
	witn := models.Witness{
		CustomerId: r.Form.Get("cust_id"),
		FirstName:  r.Form.Get("firstname"),
		LastName:   r.Form.Get("lastname"),
		Phone:      p,
		Terms:      r.Form.Get("terms"),
		Image:      witnImg,
		UserId:     userId,
	}

	data := make(map[string]interface{})
	form := forms.New(r.Form)
	form.Required("cust_id", "firstname", "lastname", "phone", "terms")
	if !form.Valid() {
		meta := models.FormMetaData{
			Message: "Edit Witness Information",
			Button:  "Edit Witness Information",
			Url:     "/admin/update-witness",
			Section: "Contract",
		}

		data["witness"] = witn
		data["metadata"] = meta

		render.Template(w, r, "witnessform.page.html", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	err = m.DB.UpdateWitness(witn)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Witness informaction could not be updated")
		http.Redirect(w, r, "/admin/add-witness", http.StatusSeeOther)
		m.App.ErrorLog.Println(err)
		return
	}

	data["witness"] = witn
	m.App.Session.Put(r.Context(), "witness", witn)
	m.App.Session.Put(r.Context(), "flash", "Witness information updated!")
	render.Template(w, r, "displayWitness.page.html", &models.TemplateData{
		Data: data,
	})
}

// ItemForm handles item form request
func (m *Repository) ItemForm(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]interface{})
	custId, _ := m.App.Session.Get(r.Context(), "customerId").(string)

	prods, err := m.DB.FetchAllProduct()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Server error, retry again")
		http.Redirect(w, r, "/add-item", http.StatusSeeOther)
		return
	}

	pt := models.PageTitle{
		Main:        "Contract Form",
		Sub:         "Contract",
		Description: "Add Item",
		PlaceHolder: "Deposit Amount",
	}

	if r.URL.Path == "/admin/edit-item" {
		itm, _ := m.App.Session.Get(r.Context(), "item").(models.Item)
		metaData := models.FormMetaData{
			Message: "Edit Selected Product",
			Button:  "Patch Product",
			Url:     "/admin/edit-item",
			Section: "Contract",
		}

		if itm.Quantity == 0 {
			itm.Quantity = 1
		}

		itm.Price *= float64(itm.Quantity)

		data["pageTitle"] = pt
		data["item"] = itm
		data["products"] = prods
		data["customerId"] = custId
		data["metadata"] = metaData
	} else {
		metaData := models.FormMetaData{
			Message: "Select Product",
			Button:  "Post Product",
			Url:     "/admin/add-item",
			Section: "Contract",
		}

		data["pageTitle"] = pt
		data["products"] = prods
		data["customerId"] = custId
		data["metadata"] = metaData
	}

	m.App.Session.Put(r.Context(), "products", prods)
	render.Template(w, r, "itemsform.page.html", &models.TemplateData{
		Form: forms.New(nil),
		Data: data,
	})
}

// PostItem handles the processing and saving of purchased items
func (m *Repository) PostItem(w http.ResponseWriter, r *http.Request) {
	userId, ok := m.App.Session.Get(r.Context(), "user_id").(int)
	data := make(map[string]interface{})
	if !ok {
		m.App.Session.Put(r.Context(), "error", "Failed to get user ID")
		http.Redirect(w, r, "/admin/add-item", http.StatusSeeOther)
		m.App.ErrorLog.Println("Failed to get user ID")
		return
	}

	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't process form")
		http.Redirect(w, r, "/admin/add-item", http.StatusSeeOther)
		log.Println(err)
		return
	}

	var price float64
	custId := r.Form.Get("cust_id")
	serial := r.Form.Get("serial")
	qty, _ := strconv.Atoi(r.Form.Get("quantity"))
	deposit, _ := strconv.ParseFloat(strings.TrimPrefix(r.Form.Get("deposit"), "₵"), 64)
	//qty, _ := strconv.Atoi(quantity)
	prods, _ := m.App.Session.Pop(r.Context(), "products").([]models.Product)

	form := forms.New(r.Form)

	form.Required("cust_id", "serial", "deposit", "quantity")
	if !form.Valid() {
		data := make(map[string]interface{})
		metaData := models.FormMetaData{
			Message: "Select Product",
			Button:  "Post Product",
			Url:     "/admin/add-item",
			Section: "Contract",
		}
		data["products"] = prods
		data["customerId"] = custId
		data["metadata"] = metaData

		m.App.Session.Put(r.Context(), "products", prods)
		m.App.Session.Put(r.Context(), "error", "Sorry, you took the wrong route!")
		render.Template(w, r, "itemsform.page.html", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	//Can be refactored
	for _, prod := range prods {
		if prod.Serial == serial {
			price = prod.Price
			break
		}
	}

	total := float64(qty) * float64(price)
	item := models.Item{
		CustomerId: custId,
		Serial:     serial,
		Price:      float64(price),
		Quantity:   int(qty),
		Deposit:    deposit,
		Balance:    total - float64(deposit),
		UserId:     userId,
	}

	prod := models.Product{
		Serial: serial,
		Units:  int32(qty),
		UserId: userId,
	}

	err = m.DB.InsertItem(item)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Failed to insert customer items purchased")
		http.Redirect(w, r, "/admin/add-item", http.StatusSeeOther)
		m.App.ErrorLog.Println(err)
		return
	}

	err = m.DB.DecreaseQuantity(prod)
	if err != nil {
		m.DB.DeleteProduct(item.Serial)
		m.App.Session.Put(r.Context(), "error", "Failed to insert customer items purchased")
		http.Redirect(w, r, "/admin/add-item", http.StatusSeeOther)
		m.App.ErrorLog.Println(err)
		return
	}

	item.Total = total
	data["item"] = item
	m.App.Session.Put(r.Context(), "item", item)
	m.App.Session.Put(r.Context(), "flash", "customer item saved!")
	render.Template(w, r, "displayItembought.page.html", &models.TemplateData{
		Data: data,
	})
}

// UpdateItem update items purchased on credit with changes by ID
func (m *Repository) UpdateItem(w http.ResponseWriter, r *http.Request) {
	userId, ok := m.App.Session.Get(r.Context(), "user_id").(int)
	data := make(map[string]interface{})
	if !ok {
		m.App.Session.Put(r.Context(), "error", "Failed to get user ID")
		http.Redirect(w, r, "/admin/edit-item", http.StatusSeeOther)
		m.App.ErrorLog.Println("Failed to get user ID")
		return
	}

	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't processing form!")
		http.Redirect(w, r, "/admin/edit-item", http.StatusSeeOther)
		m.App.ErrorLog.Println("Error parsing form")
		return
	}

	var price float64
	custId := r.Form.Get("cust_id")
	serial := r.Form.Get("serial")
	quantity := r.Form.Get("quantity")
	deposit, _ := strconv.ParseFloat(strings.TrimPrefix(r.Form.Get("deposit"), "₵"), 64)
	qty, _ := strconv.Atoi(quantity)
	prods, _ := m.App.Session.Pop(r.Context(), "products").([]models.Product)

	form := forms.New(r.Form)

	form.Required("cust_id", "serial", "deposit", "quantity")
	if !form.Valid() {
		data := make(map[string]interface{})
		metaData := models.FormMetaData{
			Message: "Select Product",
			Button:  "Patch Product",
			Url:     "/admin/edit-item",
			Section: "Contract",
		}
		data["products"] = prods
		data["customerId"] = custId
		data["metadata"] = metaData

		m.App.Session.Put(r.Context(), "products", prods)
		render.Template(w, r, "itemsform.page.html", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	//Can be refactored
	for _, prod := range prods {
		if prod.Serial == serial {
			price = prod.Price
			break
		}
	}

	total := float64(qty) * float64(price)
	item := models.Item{
		CustomerId: custId,
		Serial:     serial,
		Price:      float64(price),
		Quantity:   int(qty),
		Deposit:    float64(deposit),
		Balance:    total - float64(deposit),
		UserId:     userId,
	}

	err = m.DB.UpdateItem(item)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Purchased item could not be updated")
		http.Redirect(w, r, "/admin/edit-item", http.StatusSeeOther)
		m.App.ErrorLog.Println(err)
		return
	}

	item.Total = total
	data["item"] = item
	m.App.Session.Put(r.Context(), "flash", "customer item saved!")
	render.Template(w, r, "displayItembought.page.html", &models.TemplateData{
		Data: data,
	})
}

// ListCustomers handles request for customer history in the database
func (m *Repository) ListCustomers(w http.ResponseWriter, r *http.Request) {
	page := chi.URLParam(r, "page")
	pg, _ := strconv.Atoi(page)
	if pg == 0 {
		pg = 1
	}
	// user, ok := m.App.Session.Get(r.Context(), "user").(models.User)
	// if !ok {
	// 	m.App.Session.Put(r.Context(), "error", "You must logged in first!")
	// 	m.App.ErrorLog.Println("user could be found in the session!")
	// 	http.Redirect(w, r, "/", http.StatusSeeOther)
	// 	return
	// }
	meta := models.FormMetaData{
		Section: "Contract",
		Url:     "/admin/list-customers/1",
	}

	data := make(map[string]any)

	cust, err := m.DB.FetchCustomersByPage(pg)
	data["customers"] = cust
	data["metadata"] = meta

	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Customers cannot be fetched!")
		m.App.ErrorLog.Println("Customers cannot be fetched!", err)
		render.Template(w, r, "displaycustomers.page.html", &models.TemplateData{
			Data: data,
		})
		return
	}

	if len(cust) == 0 {
		m.App.Session.Put(r.Context(), "error", "No customer was found!")
		render.Template(w, r, "displaycustomers.page.html", &models.TemplateData{
			Data: data,
		})
		return
	}

	enterer, _ := m.DB.FetchUserById(cust[0].UserId)
	data["enterer"] = enterer

	render.Template(w, r, "displaycustomers.page.html", &models.TemplateData{
		Data: data,
	})
}

// Payments
// PaymentForm handler handles payment form request
func (m *Repository) PaymentForm(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]interface{})

	if r.URL.Path == "/admin/edit-payment" {
		pymt, _ := m.App.Session.Get(r.Context(), "payment").(models.Payments)
		metaData := models.FormMetaData{
			Section: "Contract",
			Message: "Edit Payment Information",
			Button:  "Patch Payment",
			Url:     "/admin/update-payment",
		}

		data["payment"] = pymt
		data["metadata"] = metaData
	} else {
		metaData := models.FormMetaData{
			Section: "Contract",
			Message: "Add Payment Information",
			Button:  "Post Payment",
			Url:     "/admin/pay",
		}
		data["payment"] = models.Customer{}
		data["metadata"] = metaData
	}

	render.Template(w, r, "paymentform.page.html", &models.TemplateData{
		Data: data,
		Form: forms.New(nil),
	})
}

// PostPayments handles the request for payment processing
func (m *Repository) PostPayments(w http.ResponseWriter, r *http.Request) {
	userId, ok := m.App.Session.Get(r.Context(), "user_id").(int)
	if !ok {
		m.App.Session.Put(r.Context(), "error", "Failed to get user ID")
		http.Redirect(w, r, "/admin/pay", http.StatusSeeOther)
		m.App.ErrorLog.Println("Failed to get user ID")
		return
	}
	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't process form")
		http.Redirect(w, r, "/admin/pay", http.StatusSeeOther)
		log.Println(err)
		return
	}

	amount, _ := strconv.ParseFloat(strings.TrimPrefix(r.Form.Get("payingamount"), "₵"), 64)
	data := make(map[string]interface{})

	p := models.Payments{
		CustomerId: r.Form.Get("customerId"),
		Month:      r.Form.Get("month"),
		Amount:     float64(amount),
		UserId:     userId,
	}

	form := forms.New(r.Form)
	form.Required("customerId", "month", "payingamount")
	if !form.Valid() {
		m.App.Session.Put(r.Context(), "error", "One of the fields is empty")
		metaData := models.FormMetaData{
			Section: "Contract",
			Message: "Add Payment Information",
			Button:  "Add Payment",
			Url:     "/admin/pay",
		}
		data["payment"] = p
		data["metadata"] = metaData
		render.Template(w, r, "paymentform.page.html", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	if p.Month == "The Month" {
		m.App.Session.Put(r.Context(), "error", "Choose the month of payment!")
		metaData := models.FormMetaData{
			Section: "Contract",
			Message: "Add Payment Information",
			Button:  "Add Payment",
			Url:     "/admin/pay",
		}
		data["payment"] = p
		data["metadata"] = metaData
		render.Template(w, r, "paymentform.page.html", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	err = m.DB.InsertPayment(p)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "error inserting payments! try again.")
		http.Redirect(w, r, "/admin/pay", http.StatusSeeOther)
		m.App.ErrorLog.Println(err)
		return
	}

	cust, err := m.DB.FetchCustomer(p.CustomerId)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "error fetching customer! try again.")
		http.Redirect(w, r, "/admin/pay", http.StatusSeeOther)
		m.App.ErrorLog.Println(err)
		return
	}

	if cust.Months != 0 {
		cust.Months--
	}

	err = m.DB.UpdateCustomer(cust)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "error updating customer! try again.")
		http.Redirect(w, r, "/admin/pay", http.StatusSeeOther)
		m.App.ErrorLog.Println(err)
		return
	}

	bal, _ := m.CalcCustomerDebt(p.CustomerId)

	if bal == 0.00 {
		c := models.Customer{
			CustomerId: p.CustomerId,
			Status:     "off_contract",
			UserId:     userId,
		}
		m.DB.UpdateContactStatus(c)
		data["payment"] = p
		m.App.Session.Put(r.Context(), "payment", p)
		m.App.Session.Put(r.Context(), "customerId", p.CustomerId)
		m.App.Session.Put(r.Context(), "flash", "Customer has completed payment!")
		render.Template(w, r, "displayPayment.page.html", &models.TemplateData{
			Data: data,
		})
		return
	}

	data["payment"] = p
	m.App.Session.Put(r.Context(), "payment", p)
	m.App.Session.Put(r.Context(), "customerId", p.CustomerId)
	m.App.Session.Put(r.Context(), "flash", "Customer's payment has been inserted!")
	render.Template(w, r, "displayPayment.page.html", &models.TemplateData{
		Data: data,
	})
}

// CalcCustomerDebt calculates customer debt to the enterprise
func (c *Repository) CalcCustomerDebt(customerId string) (float64, error) {
	if customerId == "" {
		return 0, errors.New("customer ID not provided")
	}

	custDebt, err := c.DB.CustomerDebt(customerId)
	if err != nil {
		return 0, errors.New("customer debt information can't be retrieved")
	}

	if len(custDebt) == 0 {
		return 0, fmt.Errorf("customer with this id: %s does not owe", customerId)
	}

	custPymt, err := c.DB.CustomerPayment(customerId)
	if err != nil {
		return 0, errors.New("customer payments information can't be retrieved")
	}

	amount := 0.00
	if len(custPymt) != 0 {
		for _, v := range custPymt {
			amount += v.Amount
		}
	}

	balance := 0.00
	for _, v := range custDebt {
		balance += v.Balance
	}

	balance -= amount
	return balance, nil
}

// ListPayments handles request for customer payment history in the database
func (m *Repository) ListPayments(w http.ResponseWriter, r *http.Request) {
	page := chi.URLParam(r, "page")
	pg, _ := strconv.Atoi(page)
	meta := models.FormMetaData{
		Section: "Contract",
		Url:     "/admin/list-payments/1",
	}

	data := make(map[string]any)

	pymt, err := m.DB.FetchPaymentsByPage(pg)

	var p []models.Payments
	for _, v := range pymt {
		v.Amount = helpers.ToDecimalPlace(v.Amount, 2)
		p = append(p, v)
	}
	data["payments"] = p
	data["metadata"] = meta
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Payments cannot be fetched!")
		m.App.ErrorLog.Println("Payments cannot be fetched!", err)
		render.Template(w, r, "displaypayments.page.html", &models.TemplateData{
			Data: data,
		})
		return
	}

	if len(pymt) == 0 {
		m.App.Session.Put(r.Context(), "error", "No payment was found!")
		render.Template(w, r, "displaypayments.page.html", &models.TemplateData{
			Data: data,
		})
		return
	}

	enterer, _ := m.DB.FetchUserById(pymt[0].UserId)
	data["enterer"] = enterer

	render.Template(w, r, "displaypayments.page.html", &models.TemplateData{
		Data: data,
	})
}

// ReceiptPage handle request for receipt generation page
func (m *Repository) ReceiptPage(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]any)
	metaData := models.FormMetaData{
		Section: "Contract",
		Message: "Add Payment Information",
		Button:  "Add Payment",
		Url:     "/admin/generate-receipt",
	}

	data["metadata"] = metaData
	render.Template(w, r, "terminateAgreeM.page.html", &models.TemplateData{
		Data: data,
		Form: forms.New(nil),
	})
}

// PurchaseForm handles item form request for purchase
func (m *Repository) PurchaseForm(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]interface{})
	custId, _ := m.App.Session.Get(r.Context(), "customerId").(string)

	prods, err := m.DB.FetchAllProduct()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Server error, retry again")
		http.Redirect(w, r, "/admin/add-purchase", http.StatusSeeOther)
		return
	}

	var p []models.Product
	for _, prod := range prods {
		prod.Price = helpers.ToDecimalPlace(prod.Price, 2)
		p = append(p, prod)
	}

	pt := models.PageTitle{
		Main:        "Purchase Form",
		Sub:         "Purchase",
		Description: "Add Purchase",
		PlaceHolder: "Enter Amount",
	}

	if r.URL.Path == "/admin/edit-purchase" {
		itm, _ := m.App.Session.Get(r.Context(), "item").(models.Item)
		metaData := models.FormMetaData{
			Message: "Edit Purchase",
			Button:  "Patch Purchase",
			Url:     "/admin/edit-purchase",
			Section: "Purchase",
		}
		data["pageTitle"] = pt
		data["item"] = itm
		data["products"] = p
		data["customerId"] = custId
		data["metadata"] = metaData
	} else {
		metaData := models.FormMetaData{
			Message: "Select Product",
			Button:  "Post Purchase",
			Url:     "/admin/add-purchase",
			Section: "Purchase",
		}
		data["pageTitle"] = pt
		data["products"] = p
		data["customerId"] = custId
		data["metadata"] = metaData
	}

	m.App.Session.Put(r.Context(), "products", prods)
	render.Template(w, r, "itemsform.page.html", &models.TemplateData{
		Form: forms.New(nil),
		Data: data,
	})
}

// Purchases
// PostPurchase handle request for the processing of purchase
func (m *Repository) PostPurchase(w http.ResponseWriter, r *http.Request) {
	userId, ok := m.App.Session.Get(r.Context(), "user_id").(int)
	data := make(map[string]interface{})
	if !ok {
		m.App.Session.Put(r.Context(), "error", "Failed to get user ID")
		http.Redirect(w, r, "/admin/add-purchase", http.StatusSeeOther)
		m.App.ErrorLog.Println("Failed to get user ID")
		return
	}

	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't process form")
		http.Redirect(w, r, "/admin/add-purchase", http.StatusSeeOther)
		m.App.ErrorLog.Println(err)
		return
	}

	serial := r.Form.Get("serial")
	quantity, _ := strconv.Atoi(r.Form.Get("quantity"))
	amount, _ := strconv.ParseFloat(r.Form.Get("amount"), 64)
	prods, _ := m.App.Session.Pop(r.Context(), "products").([]models.Product)

	form := forms.New(r.Form)

	form.Required("cust_id", "serial", "quantity", "price")
	if !form.Valid() {
		metaData := models.FormMetaData{
			Message: "Select Product",
			Button:  "Post Product",
			Url:     "/admin/add-item",
			Section: "Contract",
		}
		data["products"] = prods
		data["customerId"] = ""
		data["metadata"] = metaData

		render.Template(w, r, "itemsform.page.html", &models.TemplateData{
			Form: form,
			Data: data,
		})
	}

	p := models.Purchases{
		Serial:   serial,
		Quantity: quantity,
		Amount:   amount,
		UserId:   userId,
	}

	id, err := m.DB.InsertPurchase(p)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Error inserting purchase! try again.")
		http.Redirect(w, r, "/admin/add-purchase", http.StatusSeeOther)
		m.App.ErrorLog.Println(err)
		return
	}

	prod := models.Product{
		Serial: serial,
		Units:  int32(quantity),
		UserId: userId,
	}

	err = m.DB.DecreaseQuantity(prod)
	if err != nil {
		m.DB.DeletePurchase(id)
		m.App.Session.Put(r.Context(), "error", "Error inserting purchase! try again.")
		http.Redirect(w, r, "/admin/add-purchase", http.StatusSeeOther)
		return
	}

	m.App.Session.Put(r.Context(), "flash", "Customer's purchase is saved!")
	http.Redirect(w, r, "/admin/add-purchase", http.StatusSeeOther)
}

// ListPurchases handles request for customer purchase history in the database
func (m *Repository) ListPurchases(w http.ResponseWriter, r *http.Request) {
	page := chi.URLParam(r, "page")
	pg, _ := strconv.Atoi(page)
	if pg == 0 {
		pg = 1
	}
	meta := models.FormMetaData{
		Section: "Purchase",
		Url:     "/admin/list-purchases/1",
	}

	data := make(map[string]any)

	p, err := m.DB.FetchPurchaseByPage(pg)
	data["purchases"] = p
	data["metadata"] = meta
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Purchases cannot be fetched!")
		m.App.ErrorLog.Println("Purchases cannot be fetched!", err)
		render.Template(w, r, "displaypurchases.page.html", &models.TemplateData{
			Data: data,
		})
		return
	}

	if len(p) == 0 {
		m.App.Session.Put(r.Context(), "error", "No purchase was found!")
		render.Template(w, r, "displaypurchases.page.html", &models.TemplateData{
			Data: data,
		})
		return
	}

	var purch []models.Purchases
	for _, v := range p {
		v.Amount = helpers.ToDecimalPlace(v.Amount, 2)
		purch = append(purch, v)
	}

	enterer, _ := m.DB.FetchUserById(purch[0].UserId)

	data["purchases"] = purch
	data["enterer"] = enterer

	render.Template(w, r, "displaypurchases.page.html", &models.TemplateData{
		Data: data,
	})
}

// ListUsers handles request for users information in the database
func (m *Repository) ListUsers(w http.ResponseWriter, r *http.Request) {
	meta := models.FormMetaData{
		Section: "User",
		Url:     "/admin/list-users",
	}

	data := make(map[string]any)

	urs, err := m.DB.FetchAllUsers()
	data["users"] = urs
	data["metadata"] = meta
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Users cannot be fetched!")
		m.App.ErrorLog.Println("Users cannot be fetched!", err)
		render.Template(w, r, "displayusers.page.html", &models.TemplateData{
			Data: data,
		})
		return
	}

	if len(urs) == 0 {
		m.App.Session.Put(r.Context(), "error", "No user was found!")
		render.Template(w, r, "displayusers.page.html", &models.TemplateData{
			Data: data,
		})
		return
	}

	render.Template(w, r, "displayusers.page.html", &models.TemplateData{
		Data: data,
	})
}

// Backup and recovery
// BackupAndRecovery handles request for backing up and recovery from the database
func (m *Repository) BackupAndRecovery(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]interface{})
	var metaData models.FormMetaData
	dumpFilePath := "/home/jofosuware/Documents/OseeEA/Backups/oseeea.sql"

	if r.URL.Path == "/admin/backup" {
		metaData = models.FormMetaData{
			Section: "Backup",
			Url:     "/admin/backup",
		}
		//sourceDB := "oseeea.go"

		cmd := exec.Command("pg_dump", "-Fc", "-h", "127.0.0.1", "-U", "postgres", "oseeea.go", "-f",
			dumpFilePath)
		cmd.Env = append(os.Environ(), "PGPASSWORD=Science@1992")

		ouput, err := cmd.CombinedOutput()
		if err != nil {
			data["message"] = "internal server error, try again or reach out to the developer"
			data["metadata"] = metaData

			render.Template(w, r, "seedingform.page.html", &models.TemplateData{
				Data: data,
				Form: forms.New(nil),
			})
			return
		}
		data["message"] = "Backups is successfull!"
		m.App.InfoLog.Println(string(ouput))
	} else {
		data["message"] = "Database backup is restored!"
		metaData = models.FormMetaData{
			Section: "Backup",
			Url:     "/admin/restore",
		}

		tables, err := m.DB.ListTables()
		if err != nil {
			data["message"] = "internal server error, try again or reach out to the developer"
			data["metadata"] = metaData

			render.Template(w, r, "seedingform.page.html", &models.TemplateData{
				Data: data,
				Form: forms.New(nil),
			})
			m.App.ErrorLog.Println(err)
			return
		}

		err = m.DB.DropTables(tables)
		if err != nil {
			data["message"] = "internal server error, try again or reach out to the developer"
			data["metadata"] = metaData

			render.Template(w, r, "seedingform.page.html", &models.TemplateData{
				Data: data,
				Form: forms.New(nil),
			})
			m.App.ErrorLog.Println(err)
			return
		}

		cmd := exec.Command("pg_restore", "-d", "oseeea.go", "-h", "127.0.0.1", "-U", "postgres",
			dumpFilePath)
		cmd.Env = append(os.Environ(), "PGPASSWORD=Science@1992")

		ouput, err := cmd.CombinedOutput()
		if err != nil {
			data["message"] = "internal server error, try again or reach out to the developer"
			data["metadata"] = metaData

			render.Template(w, r, "seedingform.page.html", &models.TemplateData{
				Data: data,
				Form: forms.New(nil),
			})
		}

		m.App.InfoLog.Println(string(ouput))
	}

	render.Template(w, r, "seedingform.page.html", &models.TemplateData{
		Data: data,
		Form: forms.New(nil),
	})
}
