package comm

import (
	"github.com/garyburd/redigo/redis"
	"time"
)

var RedisPool   *redis.Pool

func InitRedis(addr, passwd string) error {
	RedisPool = &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", addr, redis.DialPassword(passwd))
		},
		MaxIdle:         100,
		MaxActive:       1000,
		IdleTimeout:     5 * time.Minute,
		Wait:            true,
	}
	return nil
}

func GetLive(liveid string) (live *pb.LiveInfo, err error) {
	live = &pb.LiveInfo{}
	conn := comm.RedisPool.Get()
	defer conn.Close()      // 如果忘了加上这一行，会导致redis连接池占满，链接得不到释放，服务被阻塞住，引起严重的bug
	bs, err := redis.Bytes(conn.Do("GET", liveid))
	if err != nil {
		return
	}
	err = proto.Unmarshal(bs, live)

	return
}

func MGetLive(liveids []string) (lives []*pb.LiveInfo, err error) {
	var keys []interface{}
	for _, liveid := range liveids {
		if len(liveid) < 0 {
			continue
		}
		keys = append(keys, liveid)
	}
	lives = make([]*pb.LiveInfo, 0)
	if len(keys) == 0 {
		return
	}
	conn := comm.RedisPool.Get()
	defer conn.Close()
	result, err := redis.ByteSlices(conn.Do("MGET", keys...))
	if err != nil {
		return
	}
  //...
  return
 }
 
 
 func GetLiveListByStatus(product, status, offset, limit int) (lives []string, err error) {
	key := fmt.Sprintf("live_list::%d::%d", product, status)
	lives = make([]string, 0)
	conn := comm.RedisPool.Get()
	defer conn.Close()
	result, rerr := redis.Strings(conn.Do("ZRANGE", key, offset, limit + offset))
	if rerr != nil {
		glog.Errorf("redis call failed, err:%v", rerr)
		return
	}
	for _, liveid := range result {
		lives = append(lives, liveid)
	}
	return
}

