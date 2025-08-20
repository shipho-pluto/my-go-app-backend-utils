package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

type (
	Config struct {
		Addr        string        `yaml:"addr"`
		Password    string        `yaml:"password"`
		User        string        `yaml:"user"`
		DB          int           `yaml:"db"`
		MaxRetries  int           `yaml:"max_retries"`
		DialTimeout time.Duration `yaml:"dial_timeout"`
		Timeout     time.Duration `yaml:"timeout"`
	}
)

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}) // init

	ctx := context.Background()
	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(pong) // ping - pong

	err = rdb.Set(ctx, "key", "value", 0).Err()
	if err != nil {
		log.Fatal(err)
	}

	val, err := rdb.Get(ctx, "key").Result()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(val) // get - set value

	err = rdb.HSet(ctx, "user:1", map[string]interface{}{
		"name": "Ivan",
		"age":  30,
	}).Err()

	user, err := rdb.HGetAll(ctx, "user:1").Result()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(user) // hash (hash map)

	rdb.RPush(ctx, "tasks", "task1", "task2", "task3") // list (list_name, list_elements...)
	tasks, _ := rdb.LRange(ctx, "tasks", 0, -1).Result()
	fmt.Println(tasks)

	rdb.SAdd(ctx, "tags", "goolang", "redis", "backend")
	tags, _ := rdb.SMembers(ctx, "tags").Result()
	fmt.Println(tags) // set (uniq value)

	rdb.Set(ctx, "temp_key", "data", 10*time.Second) // ttl value

	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	pubsub := rdb.Subscribe(ctxWithTimeout, "cannal") // pub/sub
	defer pubsub.Close()

	rdb.Publish(ctxWithTimeout, "cannal", "massage") // publish

	ch := pubsub.Channel() // subscribe
	for {
		select {
		case <-ctxWithTimeout.Done():
			return
		case msg := <-ch:
			fmt.Println(msg.Channel, msg.Payload)
		}

	}
}

func IsRateLimit(ctx context.Context, rdb redis.Client, ip string) (bool, error) { // rate limit (anti DDoS)
	key := "rate:" + ip
	count, err := rdb.Incr(ctx, key).Result()
	if err != nil {
		log.Fatal(err)
	}

	if count == 1 {
		rdb.Expire(ctx, key, time.Hour)
	}

	return count > 100, nil
}
