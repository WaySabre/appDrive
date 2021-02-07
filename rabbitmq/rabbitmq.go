package rabbitmq

import (
	"app/config"

	"github.com/streadway/amqp"
)

type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   amqp.Queue
}

// Init 初始化数据
func Init(queue_name string) (*RabbitMQ, error) {
	conf := config.GetConfAll()

	mq := new(RabbitMQ)
	url := "amqp://" + conf.MqUser + ":" + conf.MqPwd + "@" + conf.MqUrl + "/"

	var err error
	mq.conn, err = amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	mq.channel, err = mq.conn.Channel()
	if err != nil {
		return nil, err
	}

	/*err = mq.channel.ExchangeDeclare(exchange_name,amqp.ExchangeDirect,true,false,false,false,nil)
	if err != nil {
		return nil, err
	}*/

	mq.queue, err = mq.channel.QueueDeclare(queue_name, true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	return mq, nil
}

// Push 推送数据
// route_key 路由键
// message 消息
func (mq *RabbitMQ) Push(message []byte, route_key, exchange_name string) error {
	err := mq.channel.Publish(
		exchange_name,
		route_key,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         message,
		})

	return err
}

func Recive() {

}

// Close 关闭连接
func (mq *RabbitMQ) Close() error {
	err := mq.channel.Close()
	if err != nil {
		return err
	}

	err = mq.conn.Close()
	return err
}
