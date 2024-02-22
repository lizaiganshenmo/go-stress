package library

import "strings"

const (
	// protocol
	ProtocolHTTP      = "http"
	ProtocolWebSocket = "webSocket"
	ProtocolTCP       = "tcp"

	// stress
	StressMasterName = "stress_master_driver"
	StressNodeName   = "stress_work_driver"

	// snowflake
	SnowflakeWorkerID     = 1
	SnowflakeDatacenterID = 1

	SnowflakeDBWorkerID     = 2
	SnowflakeDBDatacenterID = 2

	// influxDB
	InfluxDBOrg        = "ggbond"
	InfluxDBBucket     = "testnew"
	InfluxDBMeasurment = "stress"
)

func InArrayStr(str string, arr []string) (inArray bool) {
	for _, s := range arr {
		if s == str {
			inArray = true
			break
		}
	}
	return
}

// get protocol by url
func GetURLProtocol(url string) string {
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return ProtocolHTTP
	} else if strings.HasPrefix(url, "ws://") || strings.HasPrefix(url, "wss://") {
		return ProtocolWebSocket
	}

	return ProtocolHTTP
}
