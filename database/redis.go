package database

import (
	"context"
	"fmt"
	t "time"

	r "github.com/redis/go-redis/v9"
)

func GetHolidays(db string) string {

	ctx := context.Background()

	opt, err := r.ParseURL(db)
	if err != nil {
		panic(err)
	}
	
	client := r.NewClient(opt)
	
	current_year := t.Now().Year()
	redis_key := fmt.Sprintf("holidays:%v",current_year)
	
	val, err := client.Get(ctx, redis_key).Result()
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf("%v", val)

}
