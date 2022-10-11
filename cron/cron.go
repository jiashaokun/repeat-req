package cron

import (
	"fmt"
	"github.com/go-co-op/gocron"
	"github.com/jiashaokun/repeat-req/service"
	"time"
)

func Init() {
	s := gocron.NewScheduler(time.UTC)
	fmt.Println("Crontab--------------start----------------")
	_, err := s.Cron("*/1 * * * *").Do(service.CrontabDo)
	fmt.Println("Crontab-----err------", err)
	s.StartAsync()
}
