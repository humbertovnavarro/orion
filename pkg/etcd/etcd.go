package etcd

import (
	"context"
	"errors"
	"time"

	"github.com/sirupsen/logrus"
	"go.etcd.io/etcd/clientv3"
)

var Client clientv3.Client

func init() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		panic(err)
	}
	logrus.Info("connected to etcd")
	Client = *cli
}

func Put(key string, value string) error {
	resp, err := Client.Put(Timeout(), key, value)
	logrus.Infof("%+v\n", resp)
	return err
}

func Get(key string) (out string, err error) {
	resp, err := Client.Get(Timeout(), key)

	if err != nil {
		return
	}

	if resp.Kvs == nil {
		return "", errors.New("nil response while fetching from etcd")
	}

	out = resp.Kvs[0].String()

	return
}

func Timeout() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	return ctx
}
