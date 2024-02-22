package resource

import (
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/lizaiganshenmo/GoStress/conf"
	"github.com/lizaiganshenmo/GoStress/library"
	"github.com/lizaiganshenmo/GoStress/library/utils"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// pkg globbar var
var (
	RedisCli      *redis.Client     // redis cli
	SF            *utils.Snowflake  // 用于生成node_id
	MySQLStressDB *gorm.DB          //mysql cli
	TaskSF        *utils.Snowflake  // 用于生成task_id
	InfluxDBCli   *influxdb2.Client // influxdb cli
)

func Init() {
	RedisCli = initRedis("redis_stress")
	SF = initSF(library.SnowflakeDatacenterID, library.SnowflakeWorkerID)

	MySQLStressDB = initMySQL("mysql_stress")
	TaskSF = initSF(library.SnowflakeDBDatacenterID, library.SnowflakeDBWorkerID)

	InfluxDBCli = initInfluxDB("influxdb_stress")
}

// init redis -强依赖
func initRedis(srvName string) *redis.Client {
	var err error
	var cli *redis.Client
	cli, err = conf.GetRedisClient(srvName)
	if err != nil {
		panic(err)
	}

	return cli
}

// init mysql -强依赖
func initMySQL(srvName string) *gorm.DB {
	var dsn string
	var err error
	dsn, err = conf.GetMysqlDSN(srvName)
	if err != nil {
		panic(err)
	}

	var db *gorm.DB
	db, err = gorm.Open(mysql.Open(dsn),
		&gorm.Config{
			PrepareStmt:            true,
			SkipDefaultTransaction: true, // 禁用默认事务
		},
	)
	if err != nil {
		panic(err)
	}

	return db
}

// init snowflake
func initSF(DatacenterID, WorkerID int64) *utils.Snowflake {
	var err error
	var sf *utils.Snowflake
	if sf, err = utils.NewSnowflake(DatacenterID, WorkerID); err != nil {
		panic(err)
	}

	return sf
}

// init influxdb cli
func initInfluxDB(srvName string) *influxdb2.Client {
	var err error
	var cli *influxdb2.Client

	cli, err = conf.GetInfluxDBClient(srvName)

	if err != nil {
		panic(err)
	}
	return cli
}
