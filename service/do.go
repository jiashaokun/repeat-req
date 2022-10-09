package service

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/jiashaokun/repeat-req/cache"
	"github.com/parnurzeal/gorequest"
	"strings"
	"time"
)

type RepeatReq struct {
	KeyHash string    `json:"key_hash"`
	Param   *UrlParam `json:"param"`
	Repeat  *Repeat   `json:"repeat"`
}

type UrlParam struct {
	Url             string           `json:"url"`
	Method          string           `json:"method"`
	Param           string           `json:"param"`
	RequestResponse *RequestResponse `json:"request_response"`
}

// Repeat Num 剩余次数
// Interval 剩余间隔
// NextTime 下一次请求的时间
// NextKey  下一次请求的key
type Repeat struct {
	Num      int        `json:"num"`
	Interval []int      `json:"interval"`
	NextTime *time.Time `json:"next_time"`
}

type RequestResponse struct {
	Response     string `json:"response"`
	ResponseHash string `json:"response_hash"`
}

// Set 写入
func (r *RepeatReq) Set() error {
	r.next()
	return nil
}

// 计算下一次的参数
func (r *RepeatReq) next() error {
	//计算下一个时间的key
	if r.Repeat.Num == 0 {
		return nil
	}
	//检查是否存在返回值效验
	r.nextParam()
	return nil
}

// 计算下一次的参数
func (r *RepeatReq) nextParam() error {
	//检查是否存在返回值效验
	r.Repeat.Num = r.Repeat.Num - 1
	nextMinuteNum := 1
	if len(r.Repeat.Interval) > 0 {
		//计算第一个数据的值
		nextMinuteNum = r.Repeat.Interval[0]
	}
	r.Repeat.Interval = r.Repeat.Interval[1:]
	//下一次的时间
	if r.Repeat.NextTime == nil {
		now := time.Now()
		r.Repeat.NextTime = &now
	} else {
		now := *r.Repeat.NextTime
		timeLang := time.Minute * time.Duration(nextMinuteNum)
		next := now.Add(timeLang)
		r.Repeat.NextTime = &next
	}
	r.set()
	return nil
}

// request do
func (r *RepeatReq) do() error {
	if r.Repeat.Num == 0 {
		return nil
	}
	if ok := r.request(); ok {
		return nil
	}
	r.nextParam()
	return nil
}

// 返回值 true: 继续next 、false：停止next
func (r *RepeatReq) request() bool {
	//判断是否存在返回值
	method := strings.ToUpper(r.Param.Method)
	var param map[string]interface{}
	if len(r.Param.Param) > 0 {
		if err := json.Unmarshal([]byte(r.Param.Param), &param); err != nil {
			return false
		}
	}
	var response string
	request := gorequest.New()
	switch method {
	case "POST":
		_, response, _ = request.Post(r.Param.Url).SendMap(param).End()
		break
	case "GET":
		_, response, _ = request.Get(r.Param.Url).SendMap(param).End()
		break
	}
	if len(r.Param.RequestResponse.Response) > 0 {
		paramResp := r.Param.RequestResponse.Response
		if md5.Sum([]byte(response)) == md5.Sum([]byte(paramResp)) {
			return false
		}
	}
	return true
}

//获取key
func (r *RepeatReq) set() {
	cacheKey := fmt.Sprintf("%s%s_%s", cache.BaseCacheExp, r.Param.Url, r.Param.Param)
	hashKey := md5.Sum([]byte(cacheKey))
	r.KeyHash = string(hashKey[:])

	body, _ := json.Marshal(r)

	cache.Cache.Set(r.KeyHash, string(body), time.Duration(cache.TimeCache))

	//增加下一个时段的队列 list key = 2022-01-01 01:01:01
	nextTime := *r.Repeat.NextTime
	next := nextTime.Format(cache.TimeFormat)
	nextListKey := fmt.Sprintf(cache.ListKey, next)

	listBody, ok := cache.Cache.Get(nextListKey)
	var keyList []string
	if ok {
		info := listBody.(string)
		json.Unmarshal([]byte(info), &keyList)
	}
	keyList = append(keyList, r.KeyHash)

	nextListBody, _ := json.Marshal(keyList)
	cache.Cache.Set(nextListKey, string(nextListBody), time.Duration(cache.TimeCache))
}

// CrontabDo 定时任务启动函数
func CrontabDo() {
	// cache_key
	now := time.Now().Format(cache.TimeFormat)
	keyPrefix := fmt.Sprintf(cache.ListKey, now)
	body, ok := cache.Cache.Get(keyPrefix)
	if !ok {
		return
	}
	var keyList []string
	bodyStr := body.(string)
	if err := json.Unmarshal([]byte(bodyStr), &keyList); err != nil {
		return
	}

	if len(keyList) <= 0 {
		return
	}

	// 获取数据
	go func(keyList []string) {
		for _, v := range keyList {
			infoBody, ek := cache.Cache.Get(v)
			if !ek {
				continue
			}
			infoStr := infoBody.(string)
			req := RepeatReq{}
			if reqErr := json.Unmarshal([]byte(infoStr), &req); reqErr != nil {
				continue
			}
			req.do()
		}
	}(keyList)

}
