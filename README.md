# Small Business Management App

This is a web-based application designed to help small businesses manage their inventory, customers, sales, and payments. The application is built with Go and uses a PostgreSQL database.

## Table of Contents

- [Features](#features)
- [Technologies Used](#technologies-used)
- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
- [Usage](#usage)
- [API Endpoints](#api-endpoints)
- [Database Schema](#database-schema)
- [Contributing](#contributing)
- [License](#license)

## Features

*   **User Management:** Secure user authentication with login, logout, and password reset functionality.
*   **Product Management:** Easily add, edit, delete, and search for products in your inventory.
*   **Inventory Control:** Keep track of stock levels and increase product quantities as needed.
*   **Customer & Contract Management:** Manage customer information and contracts, including witness details for agreements.
*   **Sales & Payments:**
    *   Record purchases made by customers.
    *   Handle payments for items bought on credit.
    *   Generate receipts for transactions.
*   **API:** A RESTful API to manage and retrieve data for customers, products, and payments.
*   **Backup & Recovery:** Functionality to backup and restore application data.

## Technologies Used

*   **Backend:** Go
*   **Frontend:** Go Templates, JavaScript, jQuery, Bootstrap
*   **Database:** PostgreSQL
*   **Routing:** Chi router
*   **Session Management:** SCS
*   **Database Migrations:** SQL migration files

## Project Structure

```
.
├── cmd
│   ├── api         # Main application for the API
│   └── web         # Main application for the web frontend
├── internal        # Internal application logic
│   ├── config      # Application configuration
│   ├── driver      # Database driver
│   ├── forms       # Form validation
│   ├── handlers    # HTTP handlers
│   ├── helpers     # Helper functions
│   ├── models      # Application data models
│   ├── render      # Template rendering
│   └── repository  # Database repository
├── migrations      # Database migrations
├── static          # Static assets (CSS, JS, images)
└── templates       # HTML templates
```

## Getting Started

### Prerequisites

*   Go (version 1.20 or higher)
*   PostgreSQL

### Installation

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/jofosuware/small-business-management-app.git
    cd small-business-management-app
    ```

2.  **Set up the database:**
    *   Create a PostgreSQL database.
    *   Configure your database connection in the `database.yml` file for development and production environments.
    *   Apply the database migrations located in the `/migrations` directory to set up the required tables.

3.  **Build and run the application:**
    *   The `Makefile` contains commands to build and run the application.
    *   Use `make build_back` to build the backend API.
    *   Use `make build_front` to build the frontend web server.
    *   Use `make start` to start both the backend and frontend servers.

## Usage

Once the application is running, you can access the web interface by navigating to `http://localhost:8080` in your browser. The API is available at `http://localhost:8081`.

## API Endpoints

The following are the main API endpoints available:

*   `POST /api/customer-debt/{id}`: Get the debt for a specific customer.
*   `GET /api/owing-today`: Get a list of customers who have payments due today.
*   `GET /api/list-products/{page}`: Get a paginated list of products.
*   `GET /api/list-customers/{page}`: Get a paginated list of customers.
*   `GET /api/list-payments/{page}`: Get a paginated list of payments.
*   `GET /api/list-purchases/{page}`: Get a paginated list of purchases.
*   `GET /api/expired`: Endpoint to handle system expiration (e.g., for a free trial).

## Database Schema

The database schema is defined by the SQL migration files in the `/migrations` directory. These files describe the creation and modification of the database tables.