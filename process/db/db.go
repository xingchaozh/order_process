package db

import (
	"order_process/process/util"

	"github.com/Sirupsen/logrus"
	"github.com/alphazero/Go-Redis"
)

type IDatabase interface {
	Write(stmt string, args ...interface{}) error
	Read(stmt string, recordMap *map[string]interface{}, args ...interface{}) error
	Query(stmt string, args ...interface{}) ([]map[string]interface{}, error)
}

type RedisDatabase struct {
	client redis.Client
}

var redisDB RedisDatabase

func InitDatabase(addr string, port int) error {
	spec := redis.DefaultSpec().Host(addr).Port(port)
	client, err := redis.NewSynchClientWithSpec(spec)
	if err != nil {
		return err
	}
	redisDB.client = client
	return nil
}

func Write(stmt string, args ...interface{}) error {
	key := args[0].(string)
	hashkey := args[1].(string)
	err := util.ValidateUUID(hashkey)
	if err != nil {
		return err
	}
	err = redisDB.client.Hset(key, hashkey, []byte(stmt))
	if err != nil {
		return err
	}
	return nil
}

func Read(stmt string, recordMap map[string]interface{}, args ...interface{}) error {
	key := args[0].(string)
	hashkey := args[1].(string)
	err := util.ValidateUUID(hashkey)
	if err != nil {
		return err
	}
	result, err := redisDB.client.Hget(key, hashkey)
	if err != nil {
		logrus.Error(err)
		return err
	}
	recordMap[hashkey] = result
	return nil
}

func Query(stmt string, args ...interface{}) ([]map[string]interface{}, error) {
	key := args[0].(string)
	result, err := redisDB.client.Hgetall(key)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	var maps []map[string]interface{}
	index := 0
	for _, bytes := range result {
		record := map[string]interface{}{
			string(index): string(bytes),
		}
		maps = append(maps, record)
		index++
	}
	return maps, nil
}
