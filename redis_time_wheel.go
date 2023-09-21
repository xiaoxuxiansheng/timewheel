package timewheel

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/demdxx/gocast"

	thttp "github.com/xiaoxuxiansheng/timewheel/pkg/http"
	"github.com/xiaoxuxiansheng/timewheel/pkg/redis"
	"github.com/xiaoxuxiansheng/timewheel/pkg/util"
)

type RTaskElement struct {
	Key         string            `json:"key"`
	CallbackURL string            `json:"callback_url"`
	Method      string            `json:"method"`
	Req         interface{}       `json:"req"`
	Header      map[string]string `json:"header"`
}

type RTimeWheel struct {
	sync.Once
	redisClient *redis.Client
	httpClient  *thttp.Client
	stopc       chan struct{}
	ticker      *time.Ticker
}

func NewRTimeWheel(redisClient *redis.Client, httpClient *thttp.Client) *RTimeWheel {
	r := RTimeWheel{
		ticker:      time.NewTicker(time.Second),
		redisClient: redisClient,
		httpClient:  httpClient,
		stopc:       make(chan struct{}),
	}

	go r.run()
	return &r
}

func (r *RTimeWheel) Stop() {
	r.Do(func() {
		close(r.stopc)
		r.ticker.Stop()
	})
}

func (r *RTimeWheel) AddTask(ctx context.Context, key string, task *RTaskElement, executeAt time.Time) error {
	if err := r.addTaskPrecheck(task); err != nil {
		return err
	}

	task.Key = key
	taskBody, _ := json.Marshal(task)
	_, err := r.redisClient.Eval(ctx, LuaAddTasks, 2, []interface{}{
		// 分钟级 zset 时间片
		r.getMinuteSlice(executeAt),
		// 标识任务删除的集合
		r.getDeleteSetKey(executeAt),
		// 以执行时刻的秒级时间戳作为 zset 中的 score
		executeAt.Unix(),
		// 任务明细
		string(taskBody),
		// 任务 key，用于存放在删除集合中
		key,
	})
	return err
}

func (r *RTimeWheel) RemoveTask(ctx context.Context, key string, executeAt time.Time) error {
	// 标识任务已被删除
	_, err := r.redisClient.Eval(ctx, LuaDeleteTask, 1, []interface{}{
		r.getDeleteSetKey(executeAt),
		key,
	})
	return err
}

func (r *RTimeWheel) run() {
	for {
		select {
		case <-r.stopc:
			return
		case <-r.ticker.C:
			// 每次 tick 获取任务
			go r.executeTasks()
		}
	}
}

func (r *RTimeWheel) executeTasks() {
	defer func() {
		if err := recover(); err != nil {
			// log
		}
	}()

	// 并发控制，30 s
	tctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	tasks, err := r.getExecutableTasks(tctx)
	if err != nil {
		// log
		return
	}

	// 并发执行任务
	var wg sync.WaitGroup
	for _, task := range tasks {
		wg.Add(1)
		// shadow
		task := task
		go func() {
			defer func() {
				if err := recover(); err != nil {
				}
				wg.Done()
			}()
			if err := r.executeTask(tctx, task); err != nil {
				// log
			}
		}()
	}
	wg.Wait()
}

func (r *RTimeWheel) executeTask(ctx context.Context, task *RTaskElement) error {
	return r.httpClient.JSONDo(ctx, task.Method, task.CallbackURL, task.Header, task.Req, nil)
}

func (r *RTimeWheel) addTaskPrecheck(task *RTaskElement) error {
	if task.Method != http.MethodGet && task.Method != http.MethodPost {
		return fmt.Errorf("invalid method: %s", task.Method)
	}
	if !strings.HasPrefix(task.CallbackURL, "http://") && !strings.HasPrefix(task.CallbackURL, "https://") {
		return fmt.Errorf("invalid url: %s", task.CallbackURL)
	}
	return nil
}

func (r *RTimeWheel) getExecutableTasks(ctx context.Context) ([]*RTaskElement, error) {
	now := time.Now()
	minuteSlice := r.getMinuteSlice(now)
	deleteSetKey := r.getDeleteSetKey(now)
	nowSecond := util.GetTimeSecond(now)
	score1 := nowSecond.Unix()
	score2 := nowSecond.Add(time.Second).Unix()
	rawReply, err := r.redisClient.Eval(ctx, LuaZrangeTasks, 2, []interface{}{
		minuteSlice, deleteSetKey, score1, score2,
	})
	if err != nil {
		return nil, err
	}

	replies := gocast.ToInterfaceSlice(rawReply)
	if len(replies) == 0 {
		return nil, fmt.Errorf("invalid replies: %v", replies)
	}

	deleteds := gocast.ToStringSlice(replies[0])
	deletedSet := make(map[string]struct{}, len(deleteds))
	for _, deleted := range deleteds {
		deletedSet[deleted] = struct{}{}
	}

	tasks := make([]*RTaskElement, 0, len(replies)-1)
	for i := 1; i < len(replies); i++ {
		var task RTaskElement
		if err := json.Unmarshal([]byte(gocast.ToString(replies[i])), &task); err != nil {
			// log
			continue
		}

		if _, ok := deletedSet[task.Key]; ok {
			continue
		}
		tasks = append(tasks, &task)
	}

	return tasks, nil
}

func (r *RTimeWheel) getMinuteSlice(executeAt time.Time) string {
	return fmt.Sprintf("xiaoxu_timewheel_task_{%s}", util.GetTimeMinuteStr(executeAt))
}

func (r *RTimeWheel) getDeleteSetKey(executeAt time.Time) string {
	return fmt.Sprintf("xiaoxu_timewheel_delset_{%s}", util.GetTimeMinuteStr(executeAt))
}
