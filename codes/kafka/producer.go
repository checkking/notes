package comm

import (
	"encoding/json"
	"github.com/Shopify/sarama"
	"github.com/golang/glog"
)

var Producer sarama.SyncProducer

func NewKafkaProducer(addr string) (producer sarama.SyncProducer, err error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.Return.Successes = true
	producer, err = sarama.NewSyncProducer([]string{addr}, config)
	return
}

func SendMessage(producer sarama.SyncProducer, topic string, livemsg *LiveInfoMessage) (err error) {
	msg := &sarama.ProducerMessage{}
	raw, _ := json.Marshal(livemsg)
	msg.Value = sarama.ByteEncoder(raw)
	msg.Topic = topic
	pid, offset, err := producer.SendMessage(msg)
	if err != nil {
		glog.Errorf("send message failed, err:%v", err)
	} else {
		glog.Infof("send message success, pid:%d, offset:%d", pid, offset)
	}
	return
}
