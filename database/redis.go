package database

import (
	"context"
	// j "encoding/json"
	"fmt"
	t "time"

	// "github.com/SomeoneWithOptions/api.diafestivo.co/holiday"
	r "github.com/redis/go-redis/v9"
)

func GetAllHolidaysAsJSON(db string) string {

	ctx := context.Background()

	opt, err := r.ParseURL(db)
	if err != nil {
		panic(err)
	}

	client := r.NewClient(opt)

	current_year := t.Now().Year()
	redis_key := fmt.Sprintf("holidays:%v", current_year)

	db_json, err := client.Get(ctx, redis_key).Result()
	if err != nil {
		panic(err)
	}

	return db_json

}
