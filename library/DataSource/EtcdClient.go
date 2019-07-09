package DataSource

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"sync"
	"time"
)

const (
	DefaultTimeout = time.Second * 3
)

type EtcdClient struct {
	client *clientv3.Client
}

var instance *EtcdClient
var once sync.Once

/**
 * 初始化一个单例,一般用于程序启动时
 */
func InitInstance() *EtcdClient {
	if instance == nil {
		client, err := Connect([]string{"localhost:2379"})
		if err == nil {
			instance = client
		}
	}
	return instance
}

/**
 * 获取一个单例,可以用这个不需要考虑线程安全
 */
func GetInstance() *EtcdClient {
	if instance == nil {
		instance = InitInstance()
	}
	return instance
}

/**
 * 获取一个线程安全的单例
 */
func GetSafeInstance() *EtcdClient {
	once.Do(func() {
		instance = InitInstance()
	})
	return instance
}

/**
 * 连接etcd
 */
func Connect(etcdAddrs []string) (*EtcdClient, error) {
	etcd := &EtcdClient{
	}
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   etcdAddrs,
		DialTimeout: DefaultTimeout,
	})
	if err != nil {
		return nil, err
	}
	etcd.client = cli
	return etcd, nil
}

/**
 * Set Value
 */
func (etcd *EtcdClient) Put(key string, value string) bool {
	_, err := etcd.client.Put(context.Background(), key, value)
	if err == nil {
		return true
	} else {
		return false
	}
}

/**
 * hash get,获取一个map键值对结构,对于排序的结构从ectd查出来是有序的,但map不保证有序性，所以放入map后是无序的
 */
func HGet(getResp *clientv3.GetResponse, err error) (map[string]string, error) {
	result := make(map[string]string)
	if err != nil {
		return result, err
	}

	for _, v := range getResp.Kvs {
		result[string(v.Key)] = string(v.Value)
	}
	return result, nil
}

/**
 * Get Single Key
 */
func (etcd *EtcdClient) Get(key string) (string, error) {
	getResponse, error := etcd.client.Get(context.Background(), key)
	result, err := HGet(getResponse, error)
	return result[key], err
}

/**
 * Get By prefix Mutiple Key
 */
func (etcd *EtcdClient) GetAll(prefix string) (map[string]string, error) {
	withPrefix := clientv3.WithPrefix()
	return HGet(etcd.client.Get(context.Background(), prefix, withPrefix))
}

/**
 * 获取最大键,用于获取最大ID,比如Key_001 ... Key_102 最大为Key_102
 */
func (etcd *EtcdClient) GetMaxKey(prefix string) (string, error) {
	withPrefix := clientv3.WithPrefix()
	withSort := clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend)
	withLimit := clientv3.WithLimit(1)
	resp, err := etcd.client.Get(context.Background(), prefix, withPrefix, withSort, withLimit)
	if err != nil {
		return "", err
	}

	for _, v := range resp.Kvs {
		return string(v.Key), nil
	}
	//没有数据
	return "", nil
}

/**
 * Count By prefix data
 */
func (etcd *EtcdClient) Count(prefix string) (int64, error) {
	withCount := clientv3.WithCountOnly()
	withPrefix := clientv3.WithPrefix()
	ret, err := etcd.client.Get(context.Background(), prefix, withPrefix, withCount)
	return ret.Count, err
}

/**
 * Get By prefix Limit N
 */
func (etcd *EtcdClient) GetLimit(prefix string, limit int) (map[string]string, error) {
	withPrefix := clientv3.WithPrefix()
	withLimit := clientv3.WithLimit(int64(limit))
	return HGet(etcd.client.Get(context.Background(), prefix, withPrefix, withLimit))
}

/**
 * Get By Range,Not Contains endKey,[startKey,endKey)
 */
func (etcd *EtcdClient) GetRange(startKey string, endKey string) (map[string]string, error) {
	withRange := clientv3.WithRange(endKey)
	return HGet(etcd.client.Get(context.Background(), startKey, withRange))
}

/**
 * Get By Range,Contains StartKey[startKey,N-1]
 */
func (etcd *EtcdClient) GetRangeLimit(startKey string, limit int) (map[string]string, error) {
	withLimit := clientv3.WithLimit(int64(limit))
	withFrom := clientv3.WithFromKey()
	return HGet(etcd.client.Get(context.Background(), startKey, withFrom, withLimit))
}

/**
 * Delete One
 */
func (etcd *EtcdClient) Delete(key string) (int64, error) {
	ret, err := etcd.client.Delete(context.Background(), key)
	if err != nil {
		return 0, err
	}
	return ret.Deleted, nil
}

/**
 * Delete All By Prefix
 */
func (etcd *EtcdClient) DeleteAll(prefix string) (int64, error) {
	withPrefix := clientv3.WithPrefix()
	ret, err := etcd.client.Delete(context.Background(), prefix, withPrefix)
	if err != nil {
		return 0, err
	}
	return ret.Deleted, nil
}
