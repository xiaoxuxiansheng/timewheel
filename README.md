# timewheel

<p align="center">
<img src="https://github.com/xiaoxuxiansheng/timewheel/blob/main/img/timewheel.png" height="400px/"><br/><br/>
<b>timewheel: 纯 golang 实现的时间轮框架</b>
<br/><br/>
</p>

## 📖 sdk 核心能力
- 基于 golang time ticker + 环形数组实现了单机版时间轮工具<br/><br/>
<img src="https://github.com/xiaoxuxiansheng/timewheel/blob/main/img/local_timewheel.png" height="400px"/>
- 基于 golang time ticker + redis zset 实现了分布式版时间轮工具<br/><br/>
<img src="https://github.com/xiaoxuxiansheng/timewheel/blob/main/img/zset_timewheel.png" height="400px"/>

## 💡 `原理与实现`技术博客
<a href="待补充">基于 golang 从零到一实现时间轮算法</a> <br/><br/>

## 🐧 使用示例
使用单测示例代码如下. 参见 ./time_wheel_test.go 文件<br/><br/>
- 单机版时间轮<br/><br/>
```go
func Test_timeWheel(t *testing.T) {
	timeWheel := NewTimeWheel(10, 500*time.Millisecond)
	defer timeWheel.Stop()

	timeWheel.AddTask("test1", func() {
		t.Errorf("test1, %v", time.Now())
	}, time.Now().Add(time.Second))
	timeWheel.AddTask("test2", func() {
		t.Errorf("test2, %v", time.Now())
	}, time.Now().Add(5*time.Second))
	timeWheel.AddTask("test2", func() {
		t.Errorf("test2, %v", time.Now())
	}, time.Now().Add(3*time.Second))

	<-time.After(6 * time.Second)
}
```
- redis版时间轮<br/><br/>
```go
const (
	// redis 服务器信息
	network  = "tcp"
	address  = "请输入 redis 地址"
	password = "请输入 redis 密码"
)

var (
	// 定时任务回调信息
	callbackURL    = "请输入回调地址"
	callbackMethod = "POST"
	callbackReq    interface{}
	callbackHeader map[string]string
)

func Test_redis_timeWheel(t *testing.T) {
	rTimeWheel := NewRTimeWheel(
		redis.NewClient(network, address, password),
		thttp.NewClient(),
	)
	defer rTimeWheel.Stop()

	ctx := context.Background()
	if err := rTimeWheel.AddTask(ctx, "test1", &RTaskElement{
		CallbackURL: callbackURL,
		Method:      callbackMethod,
		Req:         callbackReq,
		Header:      callbackHeader,
	}, time.Now().Add(time.Second)); err != nil {
		t.Error(err)
		return
	}

	if err := rTimeWheel.AddTask(ctx, "test2", &RTaskElement{
		CallbackURL: callbackURL,
		Method:      callbackMethod,
		Req:         callbackReq,
		Header:      callbackHeader,
	}, time.Now().Add(4*time.Second)); err != nil {
		t.Error(err)
		return
	}

	if err := rTimeWheel.RemoveTask(ctx, "test2", time.Now().Add(4*time.Second)); err != nil {
		t.Error(err)
		return
	}

	<-time.After(5 * time.Second)
	t.Log("ok")
}
```



