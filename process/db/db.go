package db

import (
	"order_process/process/util"

	"github.com/Sirupsen/logrus"
	"github.com/alphazero/Go-Redis"
)

// The database operation interface
type IDatabase interface {
	Write(stmt string, args ...interface{}) error
	Read(stmt string, recordMap *map[string]interface{}, args ...interface{}) error
	Query(stmt string, args ...interface{}) ([]map[string]interface{}, error)
}

// The definition of redis database client
type RedisDatabase struct {
	client redis.Client
}

// The instance of redis database
var redisDB RedisDatabase

// Initialize database
func InitDatabase(addr string, port int) error {
	spec := redis.DefaultSpec().Host(addr).Port(port)
	client, err := redis.NewSynchClientWithSpec(spec)
	if err != nil {
		return err
	}
	redisDB.client = client
	return nil
}

// Write operation
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

// Read operation
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

// Query operation
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
