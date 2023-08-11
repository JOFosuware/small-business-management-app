package dbrepo

import (
	"context"
	"errors"
	"time"

	"github.com/jofosuware/small-business-management-app/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func (m *postgresDBRepo) AllUsers() bool {
	return true
}

// Authenticate authenticates a user
func (m *postgresDBRepo) Authenticate(username, password string) (models.User, error) {

	u, err := m.FetchUser(username)
	if err != nil {
		return models.User{}, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return models.User{}, errors.New("incorrect password")
	} else if err != nil {
		return models.User{}, err
	}

	return u, nil
}

// InsertUser add users to the database
func (m *postgresDBRepo) InsertUser(u models.User) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var newID int

	query := `
		insert into users (first_name, last_name, user_name,
		 password, access_level, created_at, updated_at) values ($1, $2, $3, $4, $5, $6, $7) returning id
	`
	err := m.DB.QueryRowContext(ctx, query,
		u.FirstName,
		u.LastName,
		u.Username,
		u.Password,
		u.AccessLevel,
		u.CreatedAt,
		u.UpdatedAt,
	).Scan(&newID)

	if err != nil {
		return 0, err
	}

	return newID, nil
}

// FetchUser select a user with his/her username
func (m *postgresDBRepo) FetchUser(username string) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	u := models.User{}
	quary := `select 
				id, first_name, last_name, user_name, password, access_level, created_at, updated_at
		      from users where user_name = $1`

	row := m.DB.QueryRowContext(ctx, quary, username)
	err := row.Scan(
		&u.ID,
		&u.FirstName,
		&u.LastName,
		&u.Username,
		&u.Password,
		&u.AccessLevel,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		return models.User{}, err
	}

	return u, nil
}

// InsertProduct inserts product information into the database
func (m *postgresDBRepo) InsertProduct(p models.Product) (models.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var product models.Product

	query := `insert into products (serial, name, description, price, units, user_id, created_at, updated_at) 
				values ($1, $2, $3, $4, $5, $6, $7, $8) 
				returning id, serial, name, description, price, units, user_id
	`
	err := m.DB.QueryRowContext(ctx, query,
		p.Serial,
		p.Name,
		p.Description,
		p.Price,
		p.Units,
		p.UserId,
		time.Now(),
		time.Now(),
	).Scan(&product.ID, &product.Serial, &product.Name, &product.Description, &product.Price, &product.Units, &product.UserId)

	if err != nil {
		return product, err
	}

	return product, nil
}

// UpdateProduct updates product in the database by ID
func (m *postgresDBRepo) UpdateProduct(p models.Product) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		update 
			products set name = $1, description = $2, price = $3, units = $4, user_id = $5, 
			updated_at = $6 
		where 
			id = $7
	`

	_, err := m.DB.ExecContext(ctx, query,
		p.Name,
		p.Description,
		p.Price,
		p.Units,
		p.UserId,
		time.Now(),
		p.ID,
	)

	if err != nil {
		return err
	}

	return nil
}

// DecreaseQuantity updates the product's quantity by decreasing it value by value
func (m *postgresDBRepo) DecreaseQuantity(p models.Product) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		update 
			products set units = units - $1, user_id = $2, updated_at = $3 
		where 
			serial = $4
	`

	_, err := m.DB.ExecContext(ctx, query,
		p.Units,
		p.UserId,
		time.Now(),
		p.Serial,
	)

	if err != nil {
		return err
	}

	return nil
}

// IncreaseQuantity updates the product's quantity by increasing it value by value
func (m *postgresDBRepo) IncreaseQuantity(p models.Product) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		update 
			products set units = units + $1, user_id = $2, updated_at = $3 
		where 
			serial = $4
	`

	_, err := m.DB.ExecContext(ctx, query,
		p.Units,
		p.UserId,
		time.Now(),
		p.Serial,
	)

	if err != nil {
		return err
	}

	return nil
}

// FetchProduct retrieves a product with its serial number
func (m *postgresDBRepo) FetchProduct(serial string) (models.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var p models.Product

	err := m.DB.QueryRowContext(ctx,
		"select id, serial, name, description, price, units from products where serial = $1",
		serial,
	).Scan(&p.ID, &p.Serial, &p.Name, &p.Description, &p.Price, &p.Units)

	if err != nil {
		return p, err
	}

	return p, nil
}

// FetchAllProduct retrieves all product in the products database
func (m *postgresDBRepo) FetchAllProduct() ([]models.Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var p []models.Product

	rows, err := m.DB.QueryContext(ctx,
		"select * from products",
	)

	if err != nil {
		return p, err
	}

	defer rows.Close()

	for rows.Next() {
		prod := models.Product{}
		err = rows.Scan(
			&prod.ID,
			&prod.Serial,
			&prod.Name,
			&prod.Description,
			&prod.Price,
			&prod.Units,
			&prod.UserId,
			&prod.CreatedAt,
			&prod.UpdatedAt,
		)

		if err != nil {
			return p, err
		}

		p = append(p, prod)

		if err = rows.Err(); err != nil {
			return p, err
		}
	}

	return p, nil
}

// DeleteProduct removes product from the database by its serial number
func (m *postgresDBRepo) DeleteProduct(serial string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := "delete from products where serial = $1"

	_, err := m.DB.ExecContext(ctx, query, serial)
	if err != nil {
		return err
	}

	return nil
}

// FetchCustomer retrieves a customer info with its id
func (m *postgresDBRepo) FetchCustomer(customerId string) (models.Customer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var c models.Customer

	err := m.DB.QueryRowContext(ctx,
		`select 
			id, customer_id, id_type, first_name, last_name, house_address, phone, location, landmark,
			agreement, user_id, created_at, updated_at 
		from 
			customers where customer_id = $1`,
		customerId,
	).Scan(
		&c.ID,
		&c.CustomerId,
		&c.IDType,
		&c.FirstName,
		&c.LastName,
		&c.HouseAddress,
		&c.Phone,
		&c.Location,
		&c.Landmark,
		&c.Agreement,
		&c.UserId,
		&c.CreatedAt,
		&c.UpdatedAt,
	)

	if err != nil {
		return c, err
	}

	return c, nil
}

// InsertCustomer inserts customer information into the database
func (m *postgresDBRepo) InsertCustomer(c models.Customer) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `insert into customers 
				(customer_id, id_type, first_name, last_name, house_address, phone, location, 
					landmark, contract_status, agreement, user_id, created_at, updated_at) 
			  values 
			  	($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) 
	`
	_, err := m.DB.ExecContext(ctx, query,
		c.CustomerId,
		c.IDType,
		c.FirstName,
		c.LastName,
		c.HouseAddress,
		c.Phone,
		c.Location,
		c.Landmark,
		c.Status,
		c.Agreement,
		c.UserId,
		time.Now(),
		time.Now(),
	)

	if err != nil {
		return err
	}

	return nil
}

// FetchWitness retrieves a witness info with its customer's id
func (m *postgresDBRepo) FetchWitness(customerId string) (models.Witness, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var w models.Witness

	err := m.DB.QueryRowContext(ctx,
		`select 
			id, customer_id, first_name, last_name, phone, terms, user_id, created_at, updated_at 
		from 
			witness where customer_id = $1`,
		customerId,
	).Scan(
		&w.ID,
		&w.CustomerId,
		&w.FirstName,
		&w.LastName,
		&w.Phone,
		&w.Terms,
		&w.UserId,
		&w.CreatedAt,
		&w.UpdatedAt,
	)

	if err != nil {
		return w, err
	}

	return w, nil
}

// UpdateCustomer updates customer in the database by ID
func (m *postgresDBRepo) UpdateCustomer(c models.Customer) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		update 
			customers set customer_id = $1, id_type = $2, first_name = $3, last_name = $4, 
			house_address = $5, phone = $6, location = $7, landmark= $8, contract_status = $9, agreement = $10,
			user_id = $11, updated_at = $12 
		where 
			customer_id = $13 `

	_, err := m.DB.ExecContext(ctx, query,
		c.CustomerId,
		c.IDType,
		c.FirstName,
		c.LastName,
		c.HouseAddress,
		c.Phone,
		c.Location,
		c.Landmark,
		c.Status,
		c.Agreement,
		c.UserId,
		time.Now(),
		c.CustomerId,
	)

	if err != nil {
		return err
	}

	return nil
}

// UpdateContactStatus set the contract status when status changes
func (m *postgresDBRepo) UpdateContactStatus(c models.Customer) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		update customers set contract_status = $1, user_id = $2, updated_at = $3 where customer_id = $4
	`

	_, err := m.DB.ExecContext(ctx, query,
		c.Status,
		c.UserId,
		time.Now(),
		c.CustomerId,
	)

	if err != nil {
		return err
	}

	return nil
}

// InsertWitnessData inserts witness information into the database
func (m *postgresDBRepo) InsertWitnessData(w models.Witness) (models.Witness, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var witn models.Witness

	query := `insert into witness 
				(customer_id, first_name, last_name, phone, terms, user_id, created_at, updated_at) 
			  values 
			  	($1, $2, $3, $4, $5, $6, $7, $8) 
			  returning 
			  	id, customer_id, first_name, last_name, phone, terms
	`
	err := m.DB.QueryRowContext(ctx, query,
		w.CustomerId,
		w.FirstName,
		w.LastName,
		w.Phone,
		w.Terms,
		w.UserId,
		time.Now(),
		time.Now(),
	).Scan(
		&witn.ID,
		&witn.CustomerId,
		&witn.FirstName,
		&witn.LastName,
		&witn.Phone,
		&witn.Terms,
	)

	if err != nil {
		return witn, err
	}

	return witn, nil
}

// UpdateWitness updates witness in the database by ID
func (m *postgresDBRepo) UpdateWitness(w models.Witness) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		update 
			witness set first_name = $1, last_name = $2, phone = $3, terms = $4, user_id = $5, 
			updated_at = $6 
		where 
			customer_id = $7
	`

	_, err := m.DB.ExecContext(ctx, query,
		w.FirstName,
		w.LastName,
		w.Phone,
		w.Terms,
		w.UserId,
		time.Now(),
		w.CustomerId,
	)

	if err != nil {
		return err
	}

	return nil
}

// InsertItem inserts item purchased into the database
func (m *postgresDBRepo) InsertItem(itm models.Item) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `insert into 
				purchased_oncredit 
					(customer_id, serial, price, quantity, deposit, balance, user_id, created_at, updated_at) 
			  values 
			  		($1, $2, $3, $4, $5, $6, $7, $8, $9) 
			
	`
	_, err := m.DB.ExecContext(ctx, stmt,
		itm.CustomerId,
		itm.Serial,
		itm.Price,
		itm.Quantity,
		itm.Deposit,
		itm.Balance,
		itm.UserId,
		time.Now(),
		time.Now(),
	)

	if err != nil {
		return err
	}

	return nil
}

// UpdateItem updates item in the database by ID
func (m *postgresDBRepo) UpdateItem(itm models.Item) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		update 
			purchased_oncredit set serial = $1, price = $2, quantity = $3, 
			deposit = $4, balance = $5, user_id = $6, updated_at = $7
		where 
			customer_id = $8 
		AND
			serial = $9
			`

	_, err := m.DB.ExecContext(ctx, query,
		itm.Serial,
		itm.Price,
		itm.Quantity,
		itm.Deposit,
		itm.Balance,
		itm.UserId,
		time.Now(),
		itm.CustomerId,
		itm.Serial,
	)

	if err != nil {
		return err
	}

	return nil
}

// UpdateBalance updates balance of items purchased in the database by ID
func (m *postgresDBRepo) UpdateBalance(itm models.Item) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		update 
			purchased_oncredit set  
				customer_id = $1, deposit = $2, balance = $3, user_id = $4, updated_at = $5
		where 
			customer_id = $6 
		AND
			serial = $7
			`

	_, err := m.DB.ExecContext(ctx, query,
		itm.CustomerId,
		itm.Deposit,
		itm.Balance,
		itm.UserId,
		time.Now(),
		itm.CustomerId,
		itm.Serial,
	)

	if err != nil {
		return err
	}

	return nil
}

// CustomerDebt fetches custumer's balance information
func (m *postgresDBRepo) CustomerDebt(customerId string) ([]models.Item, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var custDebt []models.Item

	stmt := `SELECT 
				customer_id, serial, price, quantity, deposit, balance
			FROM
				purchased_oncredit
			WHERE
				customer_id = $1
		`
	rows, err := m.DB.QueryContext(ctx, stmt, customerId)

	if err != nil {
		return custDebt, err
	}
	defer rows.Close()

	for rows.Next() {
		var itm models.Item
		err := rows.Scan(&itm.CustomerId, &itm.Serial, &itm.Price, &itm.Quantity, &itm.Deposit, &itm.Balance)
		if err != nil {
			return custDebt, err
		}
		custDebt = append(custDebt, itm)
	}

	return custDebt, nil
}

// InsertPayment stores customer's payment information to database
func (m *postgresDBRepo) InsertPayment(p models.Payments) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		insert into
			payments
			 (customer_id, month, amount, payment_date, user_id, created_at, updated_at)
		values
			($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := m.DB.ExecContext(ctx, query,
		p.CustomerId,
		p.Month,
		p.Amount,
		time.Now(),
		p.UserId,
		time.Now(),
		time.Now(),
	)

	if err != nil {
		return err
	}

	return nil
}

// CustomerPayment fetches the payment made by a customer with his/her id
func (m *postgresDBRepo) CustomerPayment(customerId string) ([]models.Payments, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var custPayment []models.Payments

	query := `SELECT 
				customer_id, month, amount, payment_date, created_at, updated_at
			FROM
				payments
			WHERE
				customer_id = $1
		`
	rows, err := m.DB.QueryContext(ctx, query, customerId)

	if err != nil {
		return custPayment, err
	}
	defer rows.Close()

	for rows.Next() {
		var pymt models.Payments
		err := rows.Scan(&pymt.CustomerId, &pymt.Month, &pymt.Amount, &pymt.Date, &pymt.CreatedAt, &pymt.UpdatedAt)
		if err != nil {
			return custPayment, err
		}
		custPayment = append(custPayment, pymt)
	}

	return custPayment, nil
}

// InsertPurchase store the purchase made by a customer directly
func (m *postgresDBRepo) InsertPurchase(p models.Purchases) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int
	query := `insert into purchases 
				(serial, quantity, amount, user_id, created_at, updated_at) 
			  values 
			  	($1, $2, $3, $4, $5, $6) 
			  returning id
	`
	err := m.DB.QueryRowContext(ctx, query,
		p.Serial,
		p.Quantity,
		p.Amount,
		p.UserId,
		time.Now(),
		time.Now(),
	).Scan(&id)

	if err != nil {
		return id, err
	}

	return id, nil
}

// DeletePurchase removes purchases from the database by its id number
func (m *postgresDBRepo) DeletePurchase(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := "delete from purchases where id = $1"

	_, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}
