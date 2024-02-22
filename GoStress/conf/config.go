package conf

import (
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/lizaiganshenmo/GoStress/library/config"
	"github.com/olivere/elastic/v7"
	"github.com/redis/go-redis/v9"
)

func GetMysqlDSN(srvName string) (string, error) {
	return config.GetMysqlDSN(confMap, srvName)
}

func GetEsClient(srvName string) (*elastic.Client, string, error) {
	return config.GetEsClient(confMap, srvName)
}

func GetRedisClient(srvName string) (*redis.Client, error) {
	return config.GetRedisClient(confMap, srvName)
}

func GetInfluxDBClient(srvName string) (*influxdb2.Client, error) {
	return config.GetInfluxDBClient(confMap, srvName)
}
