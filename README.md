# [Source Code Validator](https://www.sourcecodevalidator.com/)


A full-stack application that validates source code files using a Golang backend and a React frontend. The project supports multiple programming languages and provides a user-friendly interface for code validation.
- `client/` – React application
- `server/` – Golang application (API)
- `docker-compose.yml` – for containerized full-stack deployment

---

## 🔧 Prerequisites

Make sure you have the following installed:

- [Node.js](https://nodejs.org/en/) – for the React frontend
- [Go](https://golang.org/dl/) – for the backend API
- [Docker & Docker Compose](https://www.docker.com/products/docker-desktop)

---

## 🚀 Running the Project

### 1. Run the React Client Locally

```bash
cd client
npm install
npm start
```

### 2. Run the Go Server Locally

```bash
cd server
go mod tidy
go run main.go
```