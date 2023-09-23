package timewheel

import (
	"context"
	"testing"
	"time"

	thttp "github.com/xiaoxuxiansheng/timewheel/pkg/http"
	"github.com/xiaoxuxiansheng/timewheel/pkg/redis"
)

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
