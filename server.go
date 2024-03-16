package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
)

// Step 1 : Global database connection
var database *sql.DB

// Step 3.1 : User struct
type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

// Step 3.2 : Err struct
type Err struct {
	Message string `json:"message"`
}

// Step 2 : login function
func login(username, password string, c echo.Context) (bool, error) {
	if username == "mark" && password == "12345" {
		return true, nil
	}
	return false, nil
}

// Step 3.0 : createUserHandler function
func createUserHandler(echoContext echo.Context) error {
	// Step 3.1
	userStruct := User{}
	err := echoContext.Bind(&userStruct)
	if err != nil {
		// Step 3.2
		return echoContext.JSON(http.StatusBadRequest, Err{Message: err.Error()})
	}
	row := database.QueryRow("INSERT INTO users (name, age) values ($1, $2)  RETURNING id", userStruct.Name, userStruct.Age)
	err = row.Scan(&userStruct.ID)
	if err != nil {
		return echoContext.JSON(http.StatusInternalServerError, Err{Message: err.Error()})
	}
	return echoContext.JSON(http.StatusCreated, userStruct)
}

// Step 4 : getUsersHandler function
func getUsersHandler(echoContext echo.Context) error {
	stagement, err := database.Prepare("SELECT id, name, age FROM users")
	if err != nil {
		return echoContext.JSON(http.StatusInternalServerError, Err{Message: "Can't prepare query all users statment:" + err.Error()})
	}
	rowResults, err := stagement.Query()
	if err != nil {
		return echoContext.JSON(http.StatusInternalServerError, Err{Message: "Can't query all users:" + err.Error()})
	}
	userStructArray := []User{}
	for rowResults.Next() {
		userStruct := User{}
		err := rowResults.Scan(&userStruct.ID, &userStruct.Name, &userStruct.Age)
		if err != nil {
			return echoContext.JSON(http.StatusInternalServerError, Err{Message: "Can't scan user:" + err.Error()})
		}
		userStructArray = append(userStructArray, userStruct)
	}
	return echoContext.JSON(http.StatusOK, userStructArray)
}

// Step 5 : getUserHandler function
func getUserHandler(echoContext echo.Context) error {
	id := echoContext.Param("id")
	stagement, err := database.Prepare("SELECT id, name, age FROM users WHERE id = $1")
	if err != nil {
		return echoContext.JSON(http.StatusInternalServerError, Err{Message: "can't prepare query user statment:" + err.Error()})
	}
	rowResult := stagement.QueryRow(id)
	userStruct := User{}
	err = rowResult.Scan(&userStruct.ID, &userStruct.Name, &userStruct.Age)
	switch err {
	case sql.ErrNoRows:
		return echoContext.JSON(http.StatusNotFound, Err{Message: "user not found"})
	case nil:
		return echoContext.JSON(http.StatusOK, userStruct)
	default:
		return echoContext.JSON(http.StatusInternalServerError, Err{Message: "can't scan user:" + err.Error()})
	}
}

func main() {
	var err error
	// Step 1
	database, err = sql.Open("postgres", "user=postgres password=Pass1234 host=127.0.0.1 port=5432 sslmode=disable")
	if err != nil {
		log.Fatal("Connect to database error : ", err)
	}
	defer database.Close()
	createTable := `CREATE TABLE IF NOT EXISTS users ( id SERIAL PRIMARY KEY, name TEXT, age INT );`
	_, err = database.Exec(createTable)
	if err != nil {
		log.Fatal("Can't create table : ", err)
	}
	echoInstance := echo.New()
	// Step 2
	echoInstance.Use(middleware.BasicAuth(login))
	//echoInstance.Use(middleware.Logger())
	//echoInstance.Use(middleware.Recover())
	// Step 3
	echoInstance.POST("/users", createUserHandler)
	// Step 4
	echoInstance.GET("/users", getUsersHandler)
	// Step 5
	echoInstance.GET("/users/:id", getUserHandler)
	log.Fatal(echoInstance.Start(":2567"))
}
