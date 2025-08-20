package handlers

import (
	"context"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/go-redis/redis/v8"
)

type (
	Card struct {
		ID   int    `json:"id" redis:"id"`
		Name string `json:"name" redis:"name"`
		Data string `json:"data" redis:"data"`
	}
)

func (c *Card) ToRerdisCard(ctx context.Context, db *redis.Client, key string) error {

	val := reflect.ValueOf(c).Elem()

	setter := func(p redis.Pipeliner) error {
		for i := 0; i < val.NumField(); i++ {
			field := val.Type().Field(i)

			tag := field.Tag.Get("redis")

			if err := p.HSet(ctx, key, tag, val.Field(i).Interface()).Err(); err != nil {
				return err
			}
		}

		if err := p.Expire(ctx, key, 30*time.Second).Err(); err != nil {
			return err
		}
		return nil
	}

	if _, err := db.Pipelined(ctx, setter); err != nil {
		return err
	}

	return nil
}

func GetCard(ctx context.Context, db *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)

		idStr := chi.URLParam(r, "id")
		if idStr == "" {
			render.Status(r, http.StatusBadRequest)
			return
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			render.Status(r, http.StatusBadRequest)
			return
		}

		card := Card{
			ID:   id,
			Name: "Test card",
			Data: "This is a test card",
		}

		if err := card.ToRerdisCard(ctx, db, idStr); err != nil {
			render.Status(r, http.StatusInternalServerError)
		}

		render.Status(r, 200)
		render.JSON(w, r, card)
	}
}

func CacheMiddleware(ctx context.Context, db *redis.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			idStr := chi.URLParam(r, "id")
			if idStr == "" {
				render.Status(r, http.StatusBadRequest)
				return
			}

			data := new(Card)
			if err := db.HGetAll(ctx, idStr).Scan(data); err == nil && (*data != Card{}) {
				render.JSON(w, r, data)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func NewCardHandler(ctx context.Context, db *redis.Client) func(r chi.Router) {
	return func(r chi.Router) {
		r.With(CacheMiddleware(ctx, db)).Get("/{id}", GetCard(ctx, db))
	}
}
