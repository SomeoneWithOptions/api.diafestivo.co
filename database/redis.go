package database

import (
	"context"
	"fmt"

	"github.com/SomeoneWithOptions/api.diafestivo.co/holiday"
	j "github.com/json-iterator/go"
	r "github.com/redis/go-redis/v9"
)

func GetAllHolidaysAsJSON(r *r.Client) (*string, error) {
	ctx := context.Background()
	c_date, _ := holiday.MakeDates(holiday.Holiday{})

	redis_key := fmt.Sprintf("holidays:%v", c_date.Year())
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
	j.Unmarshal([]byte(db_json), &holidays)

	return &holidays, nil
}
