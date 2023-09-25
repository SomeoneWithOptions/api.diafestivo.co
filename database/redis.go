package database

import (
	"context"
	"encoding/json"

	"fmt"
	t "time"

	"github.com/SomeoneWithOptions/api.diafestivo.co/holiday"
	r "github.com/redis/go-redis/v9"
)

func GetAllHolidaysAsJSON(db string) (*string, error) {

	ctx := context.Background()

	opt, errParse := r.ParseURL(db)
	if errParse != nil {
		return nil, fmt.Errorf("error parsing : %v", errParse)
	}

	client := r.NewClient(opt)

	current_year := t.Now().Year()
	redis_key := fmt.Sprintf("holidays:%v", current_year)

	db_json, errRedis := client.Get(ctx, redis_key).Result()
	if errRedis != nil {
		return nil, errRedis
	}

	return &db_json, nil
}

func GetAllHolidays(db string, year int) (*[]holiday.Holiday, error) {

	ctx := context.Background()

	opt, errParse := r.ParseURL(db)
	if errParse != nil {
		return nil, fmt.Errorf("error parsing : %v", errParse)
	}

	client := r.NewClient(opt)

	redis_key := fmt.Sprintf("holidays:%v", year)

	db_json, errRedis := client.Get(ctx, redis_key).Result()
	if errRedis != nil {
		return nil, errRedis
	}
	var holidays []holiday.Holiday
	json.Unmarshal([]byte(db_json), &holidays)

	return &holidays, nil
}
