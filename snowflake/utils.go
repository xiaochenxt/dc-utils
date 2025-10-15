package snowflake

import (
	"context"
	"github.com/dc-utils/args"
	"github.com/dc-utils/redis"
	"github.com/dc-utils/shutdown"
	"github.com/gofiber/fiber/v2/log"
	"github.com/yitter/idgenerator-go/idgen"
	"sync"
	"time"
)

var once sync.Once

func init() {
	Enable()
}

func Enable() {
	var options = idgen.NewIdGeneratorOptions(0)
	options.WorkerIdBitLength = args.GetByte("snowflake.workerIdBitLength", 6) // 默认值6，限定 WorkerId 最大值为2^6-1，即默认最多支持64个节点。
	options.SeqBitLength = args.GetByte("snowflake.seqBitLength", 6)           // 默认值6，限制每毫秒生成的ID个数。若生成速度超过5万个/秒，建议加大 SeqBitLength 到 10。
	options.BaseTime = int64(args.GetInt("snowflake.baseTime", 1657209600000)) // 如果要兼容老系统的雪花算法，此处应设置为老系统的BaseTime。
	enabled := args.GetBool("snowflake.enabled", true)
	if enabled {
		if redis.Get() != nil {
			once.Do(func() {
				var key = "dc:snowflake:workerid:list"
				var sc = `
local workerId = tonumber(ARGV[1])
            local maxWorkerIdNumber = tonumber(ARGV[2])
            local key = KEYS[1]
            for i = workerId, maxWorkerIdNumber do
              if redis.call('SISMEMBER', key, i) == 0 then
                redis.call('SADD', key, i)
                return i
              end
            end
            for i = 0, workerId - 1 do
              if redis.call('SISMEMBER', key, i) == 0 then
                redis.call('SADD', key, i)
                return i
              end
            end
            return -1
`
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				cmd, _ := redis.Get().Eval(ctx, sc, []string{key}, options.WorkerId, (1<<options.WorkerIdBitLength)-1).Int64()
				options.WorkerId = uint16(cmd)
				workerId := options.WorkerId
				log.Infof("雪花算法生成器初始化完成，机器id：%v", workerId)
				shutdown.Add(func() {
					log.Infof("雪花算法移除机器id：%v", workerId)
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()
					redis.Get().SRem(ctx, key, workerId).Val()
				})
			})
		}
	}
	idgen.SetIdGenerator(options)
}

func NextId() int64 {
	return idgen.NextId()
}
