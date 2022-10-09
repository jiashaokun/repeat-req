package cron

import (
	"github.com/go-co-op/gocron"
	"github.com/jiashaokun/repeat-req/service"
	"time"
)

func Init() {
	s := gocron.NewScheduler(time.UTC)
	s.Cron("*/1 * * * *").Do(service.CrontabDo)
	s.StartAsync()
}
