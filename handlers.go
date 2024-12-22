package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id string `json:"id" binding:"required"`
}

type RefreshParams struct {
	AccessToken  string `json:"access" binding:"required"`
	RefreshToken string `json:"refresh" binding:"required"`
}

func (app *App) AuthHandler(ctx *gin.Context) {
	var err error

	var user User

	// обработка параметров
	err = ctx.ShouldBind(&user)
	if err != nil {
		ctx.String(400, err.Error())
		return
	}

	err = uuid.Validate(user.Id)
	if err != nil {
		ctx.String(400, "invalid id")
		return
	}

	// проверка существования юзера
	err = app.conn.QueryRow(ctx, "SELECT FROM users WHERE id = $1 LIMIT 1", user.Id).Scan()
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			ctx.String(400, "unknown id")
			return
		}

		return
	}

	// генерация и обновление токенов
	acc, ref := app.updateUserTokens(ctx, ctx.ClientIP(), user.Id)

	// ответ
	ctx.JSON(200, gin.H{
		"access":  acc,
		"refresh": ref,
	})
}

func (app *App) RefreshHandler(ctx *gin.Context) {
	var err error

	var params RefreshParams

	err = ctx.ShouldBind(&params)
	if err != nil {
		ctx.String(400, err.Error())
		return
	}

	// проверить связанность токенов
	decodedRef, err := base64.StdEncoding.DecodeString(params.RefreshToken)
	if err != nil {
		ctx.String(400, "bad refresh token")
		return
	}
	if !app.tu.ValTokensPair(params.AccessToken, string(decodedRef)) {
		ctx.String(403, "bad tokens pair")
		return
	}

	// валидация и разбор access токена
	claims, err := app.tu.ParseAccessToken(params.AccessToken)
	if err != nil {
		ctx.AbortWithError(403, err)
		return
	}

	var email, refHash string
	// получение данных пользователя из бд
	err = app.conn.QueryRow(ctx, "SELECT email, refresh_token_hash FROM users WHERE id = $1 LIMIT 1", claims.UserId).Scan(&email, &refHash)
	if err != nil {
		panic(fmt.Errorf("get user data from db: %w", err))
	}

	// проверка актуальности refresh токена
	if err = bcrypt.CompareHashAndPassword([]byte(refHash), decodedRef); err != nil {
		ctx.String(403, "unknown refresh token")
		return
	}

	// отправка уведомления при смене ip
	log.Println(claims.Ip, ctx.ClientIP())
	if claims.Ip != ctx.ClientIP() {
		// TODO: работу с почтой лучше запустить отдельным потоком
		err = app.mailer.SendMail(
			email,
			fmt.Sprintf(
				"refresh request was recieved from new ip: \"%s\" (prev: \"%s\")",
				ctx.ClientIP(),
				claims.Ip,
			),
		)
		if err != nil {
			log.Println(fmt.Errorf("check ip: %w", err))
		}
	}

	acc, ref := app.updateUserTokens(ctx, ctx.ClientIP(), claims.UserId)

	// ответ
	ctx.JSON(200, gin.H{
		"access":  acc,
		"refresh": ref,
	})
}

func (app *App) updateUserTokens(ctx *gin.Context, ip string, id string) (string, string) {
	// генерация токенов
	acc, ref, err := app.tu.New(ip, id)
	if err != nil {
		panic(fmt.Errorf("gen tokens: %w", err))
	}

	// генерация bcrypt хеша для refresh токена
	hash := genBcryptHash(ref)

	// ...и занесение его в бд
	_, err = app.conn.Exec(ctx, "UPDATE users SET refresh_token_hash = $1 WHERE id = $2", hash, id)
	if err != nil {
		panic(fmt.Errorf("update user token in db: %w", err))
	}

	return acc, base64.StdEncoding.EncodeToString(ref)
}
