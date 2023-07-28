package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/jofosuware/small-business-management-app/internal/config"
	"github.com/jofosuware/small-business-management-app/internal/driver"
	"github.com/jofosuware/small-business-management-app/internal/forms"
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

	id, _, err := m.DB.Authenticate(username, password)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Invalid login credentials")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	m.App.Session.Put(r.Context(), "user_id", id)
	m.App.Session.Put(r.Context(), "flash", "Logged in successfully")
	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
}

// AddUser ceates user
func (m *Repository) AddUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}

	firstName := r.Form.Get("firstname")
	lastName := r.Form.Get("lastname")
	username := r.Form.Get("username")
	email := r.Form.Get("email")
	password := r.Form.Get("password")
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Println(err)
		return
	}

	user := models.User{
		FirstName: firstName,
		LastName:  lastName,
		Username:  username,
		Email:     email,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	form := forms.New(r.PostForm)
	form.Required("firstname", "lastname", "username", "email", "password")
	form.IsEmail("email")

	if !form.Valid() {
		log.Println("\nThe form must be filled correctly")
		return
	}

	u, err := m.DB.FetchUser(username)

	if err != nil && err.Error() != "sql: no rows in result set" {
		log.Println(err)
		return
	}

	if u.Username == username {
		log.Println("\nUser already exists")
		return
	}

	_, err = m.DB.InsertUser(user)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Printf("User: %s is inserted", user.Username)
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
		}
		data["metadata"] = metaData
	} else {
		metaData := models.FormMetaData{
			Url:     "/admin/delete-product",
			Section: "Product",
		}
		data["metadata"] = metaData
	}
	render.Template(w, r, "editproduct.page.html", &models.TemplateData{
		Data: data,
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
	form.Required("search")

	if !form.Valid() {
		m.App.Session.Put(r.Context(), "error", "form field can't be empty")
		render.Template(w, r, "editproduct.page.html", &models.TemplateData{
			Form: form,
		})
		fmt.Println("Error in form values ", err)
		return
	}

	p, err := m.DB.FetchProduct(r.Form.Get("search"))
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "No product with such serial number")
		http.Redirect(w, r, "/admin/search-product", http.StatusSeeOther)
		fmt.Println("Error Querying database ", err)
		return
	}

	m.App.Session.Put(r.Context(), "product", p)
	http.Redirect(w, r, "/admin/edit-product", http.StatusSeeOther)
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
	form.Required("search")

	if !form.Valid() {
		m.App.Session.Put(r.Context(), "error", "Invalid form data!")
		http.Redirect(w, r, "/admin/delete-product", http.StatusSeeOther)
		return
	}

	err = m.DB.DeleteProduct(r.Form.Get("search"))
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "No product with such serial number found")
		http.Redirect(w, r, "/admin/delete-product", http.StatusSeeOther)
		return
	}

	m.App.Session.Put(r.Context(), "flash", "Product deleted successfully")
	http.Redirect(w, r, "/admin/delete-product", http.StatusSeeOther)
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
	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't process form")
		http.Redirect(w, r, "/admin/add-contract", http.StatusSeeOther)
		log.Println(err)
		return
	}

	customerId := r.Form.Get("customerId")
	idType := r.Form.Get("idtype")
	firstName := r.Form.Get("firstname")
	lastName := r.Form.Get("lastname")
	hAddress := r.Form.Get("houseaddress")
	phone := r.Form.Get("phone")
	location := r.Form.Get("location")
	landmark := r.Form.Get("landmark")
	agreement := r.Form.Get("agreement")

	// do necessary convertion
	p, _ := strconv.Atoi(phone)

	c := models.Customer{
		CustomerId:   customerId,
		IDType:       idType,
		FirstName:    firstName,
		LastName:     lastName,
		Phone:        p,
		HouseAddress: hAddress,
		Location:     location,
		Landmark:     landmark,
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

	cust, err = m.DB.InsertCustomer(c)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Customer information could not be inserted!")
		http.Redirect(w, r, "/admin/add-contract", http.StatusSeeOther)
		m.App.ErrorLog.Println(err)
		return
	}

	data["customer"] = cust
	m.App.Session.Put(r.Context(), "customer", cust)
	m.App.Session.Put(r.Context(), "customerId", cust.CustomerId)
	m.App.Session.Put(r.Context(), "flash", "Customer inserted successfully")
	render.Template(w, r, "displayCustomer.page.html", &models.TemplateData{
		Data: data,
	})
}

// UpdateCustomer update customer information with changes by ID
func (m *Repository) UpdateCustomer(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
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
	agreement := r.Form.Get("agreement")

	// do necessary convertion
	p, _ := strconv.Atoi(phone)
	id, _ := strconv.Atoi(r.Form.Get("cust_id"))

	c := models.Customer{
		ID:           id,
		CustomerId:   customerId,
		IDType:       idType,
		FirstName:    firstName,
		LastName:     lastName,
		Phone:        p,
		HouseAddress: hAddress,
		Location:     location,
		Landmark:     landmark,
		Agreement:    agreement,
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
	m.App.Session.Put(r.Context(), "flash", "Customer inserted successfully")
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
		m.App.InfoLog.Println("witness data in witnessform", witn)
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

	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "can't process form")
		http.Redirect(w, r, "/admin/add-witness", http.StatusSeeOther)
		log.Println(err)
		return
	}

	p, _ := strconv.Atoi(r.Form.Get("phone"))

	witn := models.Witness{
		CustomerId: customerId,
		FirstName:  r.Form.Get("firstname"),
		LastName:   r.Form.Get("lastname"),
		Phone:      p,
		Terms:      r.Form.Get("terms"),
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
	m.App.Session.Put(r.Context(), "customer", wtn)
	m.App.Session.Put(r.Context(), "customerId", wtn.CustomerId)
	m.App.Session.Put(r.Context(), "flash", "Customer inserted successfully")
	render.Template(w, r, "displayWitness.page.html", &models.TemplateData{
		Data: data,
	})
}

// UpdateWitness update witness information with changes by ID
func (m *Repository) UpdateWitness(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "Can't processing form!")
		http.Redirect(w, r, "/admin/add-witness", http.StatusSeeOther)
		m.App.ErrorLog.Println("Error parsing form")
		return
	}

	userId := m.App.Session.Get(r.Context(), "user_id").(int)
	p, _ := strconv.Atoi(r.Form.Get("phone"))
	witn := models.Witness{
		CustomerId: r.Form.Get("cust_id"),
		FirstName:  r.Form.Get("firstname"),
		LastName:   r.Form.Get("lastname"),
		Phone:      p,
		Terms:      r.Form.Get("terms"),
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

	if r.URL.Path == "/admin/edit-item" {
		itm, _ := m.App.Session.Get(r.Context(), "item").(models.Item)
		metaData := models.FormMetaData{
			Message: "Edit Selected Product",
			Button:  "Patch Product",
			Url:     "/admin/edit-item",
			Section: "Contract",
		}
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
	quantity := r.Form.Get("quantity")
	deposit, _ := strconv.Atoi(r.Form.Get("deposit"))
	qty, _ := strconv.Atoi(quantity)
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
		render.Template(w, r, "itemsform.page.html", &models.TemplateData{
			Form: form,
			Data: data,
		})
	}

	//Can be refactored
	for _, prod := range prods {
		if prod.Serial == serial {
			price = prod.Price
			break
		}
	}

	total := float32(qty) * float32(price)
	item := models.Item{
		CustomerId: custId,
		Serial:     serial,
		Price:      float32(price),
		Quantity:   int(qty),
		Deposit:    float32(deposit),
		Balance:    total - float32(deposit),
		UserId:     userId,
	}

	err = m.DB.InsertItem(item)
	if err != nil {
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
	deposit, _ := strconv.Atoi(r.Form.Get("deposit"))
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
	}

	//Can be refactored
	for _, prod := range prods {
		if prod.Serial == serial {
			price = prod.Price
			break
		}
	}

	total := float32(qty) * float32(price)
	item := models.Item{
		CustomerId: custId,
		Serial:     serial,
		Price:      float32(price),
		Quantity:   int(qty),
		Deposit:    float32(deposit),
		Balance:    total - float32(deposit),
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

// PaymentForm handler handles payment form request
func (m *Repository) PaymentForm(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]interface{})

	if r.URL.Path == "/admin/edit-payment" {
		cust, _ := m.App.Session.Get(r.Context(), "customer").(models.Customer)
		metaData := models.FormMetaData{
			Section: "Contract",
			Message: "Edit Payment Information",
			Button:  "Update Payment",
			Url:     "/admin/update-payment",
		}

		data["customer"] = cust
		data["metadata"] = metaData
	} else {
		metaData := models.FormMetaData{
			Section: "Contract",
			Message: "Add Payment Information",
			Button:  "Add Payment",
			Url:     "/admin/pay",
		}
		data["customer"] = models.Customer{}
		data["metadata"] = metaData
	}

	render.Template(w, r, "paymentform.page.html", &models.TemplateData{
		Data: data,
		Form: forms.New(nil),
	})
}
