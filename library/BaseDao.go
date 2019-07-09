package ServiceApi

import (
	"skyway/library/DataSource"
)

type BaseDao struct {
	dataSource *DataSource.EtcdClient
}


