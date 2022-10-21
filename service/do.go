package service

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/jiashaokun/repeat-req/cache"
	"github.com/parnurzeal/gorequest"
	"net/url"
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
	nextMinuteNum := 1
	if len(r.Repeat.Interval) > 0 {
		//计算第一个数据的值
		nextMinuteNum = r.Repeat.Interval[0]
	}

	timeLang := time.Minute * time.Duration(nextMinuteNum)
	next := time.Now().Add(timeLang)
	r.Repeat.NextTime = &next

	r.set()

	return nil
}

// 计算下一次的参数 -[1,3,5]
func (r *RepeatReq) nextParam() error {
	//检查是否存在返回值效验
	r.Repeat.Num = r.Repeat.Num - 1
	if r.Repeat.Num == 0 {
		cache.Delete(r.KeyHash)
		return nil
	}
	nextMinuteNum := 1
	if len(r.Repeat.Interval) > 0 {
		if len(r.Repeat.Interval) > 1 {
			r.Repeat.Interval = r.Repeat.Interval[1:]
		}
		//计算第一个数据的值
		nextMinuteNum = r.Repeat.Interval[0]
	}
	//下一次的时间
	if r.Repeat.NextTime == nil {
		timeLang := time.Minute * time.Duration(1)
		now := time.Now().Add(timeLang)
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

// 返回值 true: 继续next 、false：停止next
func (r *RepeatReq) request() {
	//判断是否存在返回值
	method := strings.ToUpper(r.Param.Method)
	var response string
	request := gorequest.New()
	switch method {
	case "POST":
		_, response, _ = request.Post(r.Param.Url).Type("multipart").Send(r.Param.Param).End()
		break
	case "GET":
		requestUrl := r.Param.Url
		param := make(map[string]interface{})
		if len(r.Param.Param) > 0 {
			json.Unmarshal([]byte(r.Param.Param), &param)
			urlP := url.Values{}
			for k, v := range param {
				urlP.Set(k, fmt.Sprintf("%v", v))
			}
			paramUrl := urlP.Encode()
			requestUrl = fmt.Sprintf("%s?%s", r.Param.Url, paramUrl)
		}

		_, response, _ = request.Get(requestUrl).End()
		break
	}
	if len(r.Param.RequestResponse.Response) > 0 {
		paramResp := r.Param.RequestResponse.Response
		respByte := md5.Sum([]byte(response))
		respStr := fmt.Sprintf("%x", respByte)
		paramRespByte := md5.Sum([]byte(paramResp))
		paramRespStr := fmt.Sprintf("%x", paramRespByte)

		if respStr == paramRespStr {
			r.Repeat.Num = 0
		}
	}
	if r.Repeat.Num == 0 {
		return
	}
	r.nextParam()
	return
}

//获取key
func (r *RepeatReq) set() {
	cacheKey := fmt.Sprintf("%s%s_%s", cache.BaseCacheExp, r.Param.Url, r.Param.Param)
	hashKey := md5.Sum([]byte(cacheKey))

	r.KeyHash = fmt.Sprintf("%x", hashKey)
	body, _ := json.Marshal(r)
	cache.Set(r.KeyHash, string(body))

	//增加下一个时段的队列 list key = 2022-01-01 01:01:01
	nextTime := *r.Repeat.NextTime
	next := nextTime.Format(cache.TimeFormat)
	nextListKey := fmt.Sprintf(cache.ListKey, next)

	listBody := cache.Get(nextListKey)
	var keyList []string
	info := string(listBody)
	json.Unmarshal([]byte(info), &keyList)

	keyList = append(keyList, r.KeyHash)
	nextListBody, _ := json.Marshal(keyList)

	cache.Set(nextListKey, string(nextListBody))
}

// CrontabDo 定时任务启动函数
func CrontabDo() {
	// 删除上一分钟的list
	upTime := time.Now().Add(-time.Minute).Format(cache.TimeFormat)
	upPrefix := fmt.Sprintf(cache.ListKey, upTime)
	go cache.Delete(upPrefix)

	// cache_key
	now := time.Now().Format(cache.TimeFormat)
	keyPrefix := fmt.Sprintf(cache.ListKey, now)

	data := cache.Get(keyPrefix)

	var keyList []string
	bodyStr := string(data)
	if err := json.Unmarshal([]byte(bodyStr), &keyList); err != nil {
		return
	}

	if len(keyList) <= 0 {
		return
	}

	// 获取数据
	go func(keyList []string) {
		for _, v := range keyList {
			infoBody := cache.Get(v)
			infoStr := string(infoBody)
			req := RepeatReq{}
			if reqErr := json.Unmarshal([]byte(infoStr), &req); reqErr != nil {
				continue
			}
			req.request()
		}
	}(keyList)

}
