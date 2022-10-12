# Repeat Req 请求重试，解放你重复造轮子的双手

请求重试，使用本地缓存，缓存有效期为 1 年，如果 1 年都不会重启的项目，不建议使用。
本功能将根据传递的参数，自动发起http请求，目前只支持 get/post 。

## 支持
1. 支持重复发送请求次数
2. 支持重复发送请求间隔时间（分钟）
3. 支持匹配返回值，匹配成功后不再发送

### Demo
```go
//使用gin框架举例
func main() {
	//在这里初始化
    repeat_req.Init()
    r := gin.Default()

    r.GET("/ping", ttt)
    r.Run(":8801") // listen and serve o
}

func ttt(c *gin.Context) {
    
	info := map[string]int{
		"a": 111,
	}
	b, _ := json.Marshal(info)
	bi := string(b)
    
    req := repeat_req.Repeat{
        Url:      "http://127.0.0.1:8801/to",
        Method:   "GET",
        Num:      3,    //重试次数
        Interval: []int{1, 1},  //间隔时间（数量小于次数时按 1 分钟）
        Response: bi, //该值不为空时，匹配返回值，成功后不再发送请求
    }
    err := req.Do()
}
```
