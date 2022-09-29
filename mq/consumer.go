package mq

import "log"

var done chan bool

// StartConsume: 开始监听队列，接收消息
func StartConsume(qName, cName string, callback func(msg []byte) bool) {
	//1. 通过channel.Consume获取消息信道
	msgs, err := channel.Consume(
		qName,
		cName,
		true, //自动应答
		false, // 非唯一的消费者
		false, // rabbitMQ只能设置为false
		false, // noWait, false表示会阻塞直到有消息过来
		nil,
	)
	if err != nil {
		log.Println(err.Error())
		return
	}
	done = make(chan bool)
	go func() {
		//2. 循环获取队列的消息
		for msg := range msgs {
			//3. 调用callback方法来处理新的消息
			procssErr := callback(msg.Body)
			if procssErr{
				// TODO: 将任务写到错误队列，待后续处理
			}
		}
	}()
	// 接收done的信号, 没有信息过来则会一直阻塞，避免该函数退出
	<-done

	// 关闭通道
	channel.Close()
}

// StopConsume : 停止监听队列
func StopConsume() {
	done <- true
}