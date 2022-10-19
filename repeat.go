package repeat_req

import (
	"encoding/json"
	"github.com/jiashaokun/repeat-req/cache"
	"github.com/jiashaokun/repeat-req/cron"
	"github.com/jiashaokun/repeat-req/service"
)

func Init() {
	cache.InitCache()
	cron.Init()
}

// Repeat Url 需要请求的地址
// Repeat Param 请求参数
// Repeat Method 目前只有 [post/get] 默认get
// Repeat Response 请求返回校验值，举例：上述请求返回 {"code":200},若请求遇到同等值则不再请求，该值为空，则不参与校验
// Repeat Num 请求次数
// Repeat Interval 选填（不填默认分发间隔时间 1 分钟） 间隔时间，单位：分钟,间隔时间的元素个数必须等于字段 Num 的值；举例：[1,2,3]，第一次是 1 分钟后开始请求，第二次是当第一次请求完毕后间隔的时间也就是如果第一次请求失败，与第一次间隔 2 分钟后会请求第二次，以此类推，第二次若请求失败，与第二次间隔 3 分钟后请求第三次
type Repeat struct {
	Url      string                 `json:"url"`
	Param    map[string]interface{} `json:"param"`
	Method   string                 `json:"method"`
	Response string                 `json:"response"`
	Num      int                    `json:"num"`
	Interval []int                  `json:"interval"`
}

func (c *Repeat) Do() error {
	resp := service.RequestResponse{
		Response: c.Response,
	}
	param := service.UrlParam{
		Url:             c.Url,
		Method:          c.Method,
		RequestResponse: &resp,
	}

	//param 函数 json
	if len(c.Param) > 0 {
		paramByte, _ := json.Marshal(c.Param)
		param.Param = string(paramByte)
	}

	repeat := service.Repeat{
		Num:      c.Num,
		Interval: c.Interval,
	}
	info := service.RepeatReq{
		Param:  &param,
		Repeat: &repeat,
	}
	if err := info.Set(); err != nil {
		return err
	}
	return nil
}
