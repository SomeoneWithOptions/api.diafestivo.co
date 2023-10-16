package database

import (
	"context"
	"encoding/json"

	"fmt"
	t "time"

	"github.com/SomeoneWithOptions/api.diafestivo.co/holiday"
	r "github.com/redis/go-redis/v9"
)

func GetAllHolidaysAsJSON(r *r.Client) (*string, error) {
	ctx := context.Background()
	current_year := t.Now().Year()
	redis_key := fmt.Sprintf("holidays:%v", current_year)
	db_json, errRedis := r.Get(ctx, redis_key).Result()
	if errRedis != nil {
		return nil, errRedis
	}

	return &db_json, nil
}

func GetAllHolidays(r *r.Client, year int) (*[]holiday.Holiday, error) {
	ctx := context.Background()
	redis_key := fmt.Sprintf("holidays:%v", year)
	db_json, errRedis := r.Get(ctx, redis_key).Result()
	if errRedis != nil {
		return nil, errRedis
	}
	var holidays []holiday.Holiday
	json.Unmarshal([]byte(db_json), &holidays)

	return &holidays, nil
}
