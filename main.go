package main

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

type App struct {
	// env map[string]string
	conn   *pgx.Conn
	mailer *Mailer
	tu     *TokensUtils
}

func main() {
	env, err := godotenv.Read()
	if err != nil {
		log.Fatal("cant read .env: %w", err)
	}

	secret, avail := env["JWT_SECRET"]
	if !avail {
		log.Fatal("Can't start: no JWT_SECRET")
	}
	tu := &TokensUtils{
		secret: []byte(secret),
	}

	conn, err := pgx.Connect(context.Background(), env["DATABASE_URL"])
	if err != nil {
		log.Fatal(fmt.Errorf("cant connect to pg: %w", err))
	}

	mailerPort, err := strconv.Atoi(env["SMTP_PORT"])
	if err != nil {
		panic(err)
	}
	mailer := NewMailer(
		env["SMTP_HOST"],
		mailerPort,
		env["SMTP_LOGIN"],
		env["SMTP_PASS"],
		env["SMTP_FROM"],
	)

	app := App{
		conn,
		mailer,
		tu,
	}

	r := gin.Default()
	r.POST("/auth", app.AuthHandler)
	r.POST("/refresh", app.RefreshHandler)

	r.Run("0.0.0.0:8080")
}
