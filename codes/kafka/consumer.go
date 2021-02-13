package comm

import (
	"encoding/json"
	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
	"github.com/golang/glog"
	"time"
)

func NewKafkaConsumer(addr, topic, groupId string) (consumer *cluster.Consumer, err error) {
	config := cluster.NewConfig()
	config.Group.Return.Notifications = true
	config.Consumer.Offsets.CommitInterval = time.Second
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	consumer, err = cluster.NewConsumer([]string{addr}, groupId, []string{topic}, config)
	return
}


func processMessage(lm *comm.LiveInfoMessage) (err error) {
	glog.Infof("processMessage, action:%d, liveid:%d", lm.Action, lm.Liveid)
	switch lm.Action {
	case comm.MESSAGE_TYPE_ADD, comm.MESSAGE_TYPE_UPDATE:
    glog.Infof("a add or update message")
		break
	case comm.MESSAGE_TYPE_DELETE:
		// not implemented yet
		glog.Infof("a delete message")
		break
	default:
		glog.Errorf("unsupported type:%d", int(lm.Action))
		break
	}
	return
}

func liveMessageRoutine() {

	go func() {
		glog.Infof("start message consumer...")
		consumer, err := comm.NewKafkaConsumer(config.Cfg.KafkaAddr,
			config.Cfg.LiveMessageTopic, config.Cfg.LiveMessageConsumerGroup)
		if err != nil {
			glog.Errorf("NewKafkaConsumer failed, err:%v", err)
			return
		}
		ticker := time.NewTicker(time.Second * 10)
		defer ticker.Stop()
		defer consumer.Close()
		for {
			select {
			case msg := <- consumer.Messages():
				glog.Infof("got a message, topic:%s, offset:%d", msg.Topic, msg.Offset)
				consumer.MarkOffset(msg, "")
				var lm comm.LiveInfoMessage
				err := json.Unmarshal(msg.Value, &lm)
				if err != nil {
					glog.Errorf("Unmarshal failed, err:%v", err)
				}
				err = processMessage(&lm)
				if err != nil {
					glog.Errorf("processMessage failed, err:%v", err)
				}
			case noti := <- consumer.Notifications():
				glog.Infof("consumer notification:%v", noti)
			case cerr := <- consumer.Errors():
				glog.Errorf("consumer error found:%v", cerr)
			case <- ticker.C:
				glog.Infof("consumer ticker ticked..")
			case stop := <- messageConsumeChan:
				if stop {
					glog.Errorf("got stop chan")
					return
				}
			}
		}
	}()
}
