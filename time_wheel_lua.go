package timewheel

const (
	// 1 添加任务时，如果存在删除 key 的标识，则将其删除
	// 添加任务时，根据时间（所属的 min）决定数据从属于哪个分片{}
	LuaAddTasks = `
	   local zsetKey = KEYS[1]
	   local deleteSetKey = KEYS[2]
	   local score = ARGV[1]
	   local task = ARGV[2]
	   local taskKey = ARGV[3]
	   redis.call('srem',deleteSetKey,taskKey)
	   return redis.call('zadd',zsetKey,score,task)
	`

	// 2 删除任务时，将删除 key 的标识置为 true
	LuaDeleteTask = `
	   local deleteSetKey = KEYS[1]
	   local taskKey = ARGV[1]
	   redis.call('sadd',deleteSetKey,taskKey)
	   local scnt = redis.call('scard',deleteSetKey)
	   if (tonumber(scnt) == 1)
	   then
	       redis.call('expire',deleteSetKey,120)
       end
	   return scnt
	`

	// 3 执行任务时，通过 zrange 操作取回所有不存在删除 key 标识的任务
	LuaZrangeTasks = `
	   local zsetKey = KEYS[1]
	   local deleteSetKey = KEYS[2]
	   local score1 = ARGV[1]
	   local score2 = ARGV[2]
	   local deleteSet = redis.call('smembers',deleteSetKey)
	   local targets = redis.call('zrange',zsetKey,score1,score2,'byscore')
	   redis.call('zremrangebyscore',zsetKey,score1,score2)
	   local reply = {}
	   reply[1] = deleteSet
	   for i, v in ipairs(targets) do
	       reply[#reply+1]=v
	   end
       return reply
	`
)
