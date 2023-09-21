package timewheel

import (
	"context"
	"net/http"
	"testing"
	"time"

	thttp "github.com/xiaoxuxiansheng/timewheel/pkg/http"
	"github.com/xiaoxuxiansheng/timewheel/pkg/redis"
	"github.com/xiaoxuxiansheng/timewheel/pkg/util"
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
	network  = "tcp"
	address  = "43.138.61.23:6379"
	password = "19951212"
)

func Test_redis_timeWheel(t *testing.T) {
	rTimeWheel := NewRTimeWheel(
		redis.NewClient(network, address, password),
		thttp.NewClient(),
	)
	defer rTimeWheel.Stop()

	ctx := context.Background()
	if err := rTimeWheel.AddTask(ctx, "test1", &RTaskElement{
		CallbackURL: "http://localhost:8080/ping",
		Method:      http.MethodPost,
	}, time.Now().Add(time.Second)); err != nil {
		t.Error(err)
		return
	}

	if err := rTimeWheel.AddTask(ctx, "test2", &RTaskElement{
		CallbackURL: "http://localhost:8080/ping",
		Method:      http.MethodPost,
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

func Test_getTimeMinuteStr(t *testing.T) {
	t.Error(util.GetTimeMinuteStr(time.Now()))

}
