package config

import (
	"context"
	"fmt"
	"strings"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/mitchellh/mapstructure"
	"github.com/olivere/elastic/v7"
	"github.com/redis/go-redis/v9"
)

// get mysqlDSN by map conf
func GetMysqlDSN(confMap map[string]interface{}, srvName string) (string, error) {
	val, ok := confMap[srvName]
	if !ok {
		return "", fmt.Errorf("unkonwn srvname: %s", srvName)
	}

	var mc MySQLConf
	err := mapstructure.Decode(val, &mc)
	if err != nil {
		return "", err
	}

	dsn := strings.Join([]string{mc.MySQL.Username, ":",
		mc.MySQL.Password, "@tcp(", mc.MySQL.Addr, ")/", mc.MySQL.Database, "?charset=" + mc.MySQL.Charset + "&parseTime=true"}, "")
	return dsn, nil
}

// get esClient by map conf
func GetEsClient(confMap map[string]interface{}, srvName string) (*elastic.Client, string, error) {
	val, ok := confMap[srvName]
	if !ok {
		return nil, "", fmt.Errorf("unkonwn srvname: %s", srvName)
	}

	var ec EsConf
	err := mapstructure.Decode(val, &ec)
	if err != nil {
		return nil, "", err
	}

	client, err := elastic.NewClient(
		elastic.SetURL(fmt.Sprintf("http://%s:%s", ec.ES.Host, ec.ES.Port)),
		elastic.SetBasicAuth(ec.ES.UserName, ec.ES.Password),
		elastic.SetSniff(false),
	)

	return client, ec.ES.Host, err

}

// get redis client by map conf
func GetRedisClient(confMap map[string]interface{}, srvName string) (*redis.Client, error) {
	val, ok := confMap[srvName]
	if !ok {
		return nil, fmt.Errorf("unkonwn srvname: %s", srvName)
	}

	var redisConf RedisConf
	err := mapstructure.Decode(val, &redisConf)
	if err != nil {
		return nil, err
	}

	RedisClient := redis.NewClient(&redis.Options{
		Addr:     redisConf.Redis.Addr,
		Password: redisConf.Redis.Password,
	})
	_, err = RedisClient.Ping(context.TODO()).Result()
	// if err != nil {
	// 	panic(err)
	// }

	return RedisClient, err

}

// get rabbitMQ url by map conf
func GetRabbitMQUrl(confMap map[string]interface{}, srvName string) (string, error) {
	val, ok := confMap[srvName]
	if !ok {
		return "", fmt.Errorf("unkonwn srvname: %s", srvName)
	}

	var rc RabbitMQConf
	err := mapstructure.Decode(val, &rc)
	if err != nil {
		return "", err
	}

	url := strings.Join([]string{"amqp://", rc.RabbitMQ.Username, ":", rc.RabbitMQ.Password, "@", rc.RabbitMQ.Addr, "/"}, "")
	return url, nil
}

// get influxdb client
func GetInfluxDBClient(confMap map[string]interface{}, srvName string) (*influxdb2.Client, error) {
	val, ok := confMap[srvName]
	if !ok {
		return nil, fmt.Errorf("unkonwn srvname: %s", srvName)
	}

	var ic InfluxDBConf
	err := mapstructure.Decode(val, &ic)
	if err != nil {
		return nil, err
	}

	// cli := influxdb2.NewClientWithOptions(ic.InfluxDB.Addr, ic.InfluxDB.Token, influxdb2.DefaultOptions().SetBatchSize(20))
	cli := influxdb2.NewClient(ic.InfluxDB.Addr, ic.InfluxDB.Token)

	if err != nil {
		return nil, err
	}

	return &cli, nil

}
