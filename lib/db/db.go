package db

import (
	"github.com/Sirupsen/logrus"
	"github.com/alphazero/Go-Redis"
	"order_process/lib/util"
)

type IDatabase interface {
	Write(stmt string, recordMap map[string]interface{}, args ...interface{}) error
	Read(stmt string, recordMap *map[string]interface{}, args ...interface{}) error
}

type RedisDatabase struct {
	client redis.Client
}

var redisDB RedisDatabase

func InitDatabase(addr string, port int) error {
	spec := redis.DefaultSpec().Host(addr).Port(port)
	client, err := redis.NewSynchClientWithSpec(spec)
	redisDB.client = client
	return err
}

func Write(stmt string, recordMap map[string]interface{}, args ...interface{}) error {
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

func Read(stmt string, recordMap *map[string]interface{}, args ...interface{}) error {
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
	(*recordMap)[hashkey] = result
	return nil
}

/*
import (
	"order_process/lib/util"
)

var orderRecords map[string]interface{}

func InitDatabase() {
	orderRecords = make(map[string]interface{})
}

func Write(stmt string, recordMap map[string]interface{}, args ...interface{}) error {
	id := args[0].(string)
	err := util.ValidateUUID(id)
	if err != nil {
		return err
	}
	orderRecords[id] = recordMap
	return nil
}

func Read(stmt string, recordMap *map[string]interface{}, args ...interface{}) error {
	id := args[0].(string)
	err := util.ValidateUUID(id)
	if err != nil {
		return err
	}
	if _, ok := orderRecords[id].(map[string]interface{}); ok {
		*recordMap = orderRecords[id].(map[string]interface{})
	}
	return nil
}
*/
