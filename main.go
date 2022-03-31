package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/go-redis/redis/v8"
	"gopkg.in/gomail.v2"
)

type User struct {
	id    int    `json:"id"`
	Email string `json"email`
	Name  string `json"name`
}

func connect() *sql.DB {
	db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/db_latihan_tools?parseTime=true&loc=Asia%2FJakarta")
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func main() {
	// Connect to redis
	s := gocron.NewScheduler(time.UTC)
	s.Every(30).Seconds().Do(func() {
		go SendEmail()
		time.Sleep(500 * time.Millisecond)

	}) // Every 30 seconds
	s.StartAsync()
	s.StartBlocking()
}

//mail
func SendEmail() {

	var list []User = getUser()
	d := gomail.NewDialer("smtp.example.com", 587, "stevianianggila60@gmail.com", "NakNik919")

	m := gomail.NewMessage()
	for _, r := range list {
		m.SetHeader("From", "stevianianggila60@gmail.com")
		m.SetAddressHeader("To", r.Email, r.Name)
		m.SetHeader("Subject", "Newsletter #1")
		m.SetBody("text/html", fmt.Sprintf("Hello %s!", r.Name))
		if err := d.DialAndSend(m); err != nil {
			fmt.Print(err)
			panic(err)
		}
		m.Reset()
	}
}

// func SendMail() {
// 	m := gomail.NewMessage()
// 	m.SetHeader("From", "from@example.com")
// 	m.SetHeader("To", "to@example.com")
// 	m.SetHeader("Subject", "Hello!")
// 	m.SetBody("text/plain", "Hello!")

// 	d := gomail.Dialer{Host: "localhost", Port: 587}
// 	if err := d.DialAndSend(m); err != nil {
// 		panic(err)
// 	}
// }

//redis
func getUser() []User {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	var ctx = context.Background()

	value, err := client.Get(ctx, "users").Result()
	if err != nil {
		log.Println("Gagal melakukan get")
		log.Println(err)
		return nil
	}

	var users []User
	_ = json.Unmarshal([]byte(value), &users)

	return users
}

func setUser(users []User) {
	converted, err := json.Marshal(users)
	if err != nil {
		log.Println(err)
		return
	}

	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	var ctx = context.Background()

	err = client.Set(ctx, "users", converted, 0).Err()
	if err != nil {
		log.Println("Gagal melakukan set")
		log.Println(err)
		return
	} else {
		log.Println("Cache didapat")
	}
}

func GetAllUsers(w http.ResponseWriter, r *http.Request) {

	var users []User

	if users == nil {
		db := connect()
		defer db.Close()

		query := "Select id, email, nama from users"
		rows, err := db.Query(query)
		if err != nil {
			log.Println(err)
			log.Println(w, 400, "Query Error")
			return
		}

		var user User

		for rows.Next() {
			if err := rows.Scan(&user.id, &user.Email, &user.Name); err != nil {
				log.Println(err.Error())
				log.Println(w, 400, "gagal get")
				return
			} else {
				users = append(users, user)
			}
		}
		setUser(users)
	}
}
