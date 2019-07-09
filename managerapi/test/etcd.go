package main

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"log"
	"skyway/library/DataSource"
	"skyway/managerapi/dao"
	"skyway/managerapi/model"
	"strconv"
	"time"
)

func test1() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		fmt.Println("connect failed, err:", err)
		return
	}

	fmt.Println("connect succ")
	defer cli.Close()

	for i := 1000; i < 1020; i++ {
		cli.Put(context.Background(), "api_"+strconv.Itoa(i), "value_"+strconv.Itoa(i))
	}

	withCount := clientv3.WithCountOnly()
	withPrefix := clientv3.WithPrefix()
	resp, _ := cli.Get(context.Background(), "api", withPrefix, withCount)
	log.Println(resp.Count)
	log.Println(resp)

	withRange := clientv3.WithRange("api_1011")
	withLimit := clientv3.WithLimit(4)
	resp2, _ := cli.Get(context.Background(), "api_", withRange, withLimit)
	log.Println(resp2.Count)
	for _, item := range resp2.Kvs {
		log.Println(item)
	}
}

func test2() {
	client := DataSource.InitInstance()

	api1 := model.NewApi()
	api1.ApiId = 1001
	api1.ApiDescription = "test1111111"
	api1.ApiName = "test1111111name"

	api2 := model.NewApi()
	api2.ApiId = 1002
	api2.ApiDescription = "test1112222"
	api2.ApiName = "test1112222name"

	apiDao := DAO.NewApiDao()
	_, _ = apiDao.DelAllApis()

	apiDao.RegisterApi(api1)
	apiDao.RegisterApi(api2)

	apiMap, _ := apiDao.GetApis()
	maxId, _ := apiDao.GetMaxId()

	log.Println(apiMap)
	log.Println(maxId)

	client.Put("api/11", "xxxx")
	client.Put("test/1004", "xxxxx1004")
	client.Put("test/1001", "xxxxx1001")
	client.Put("test/1007", "xxxxx1007")
	client.Put("test/1002", "xxxxx1002")

	retVal, _ := client.Get("test/1004")
	log.Println(retVal)

	ret1, _ := client.Count("test")
	log.Println(ret1)

	ret, _ := client.GetAll("test")
	log.Println(ret)

	ret, _ = client.GetRange("test/1002", "test/1007")
	log.Println(ret)

	ret, _ = client.GetLimit("test", 4)
	log.Println(ret)

	ret, _ = client.GetRangeLimit("test/1004", 4)
	log.Println(ret)
}

func main() {
	test2()
}
