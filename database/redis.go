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
	holidays_json, err := r.Get(ctx, redis_key).Result()

	if err != nil {
		return nil, err
	}

	return &holidays_json, nil
}

func GetAllHolidays(r *r.Client, year int) (*[]holiday.Holiday, error) {
	ctx := context.Background()
	redis_key := fmt.Sprintf("holidays:%v", year)

	fmt.Println("Redis Key: ",redis_key)

	err := r.Ping(ctx).Err()

	if err != nil {
		fmt.Println("BAD DB")
		return nil, err
	}

	fmt.Println("OK DB")

	holidaysJSON, err := r.Get(ctx, redis_key).Result()

	if err != nil {
		return nil, err
	}

	var holidays []holiday.Holiday
	j.Unmarshal([]byte(holidaysJSON), &holidays)

	return &holidays, nil
}

