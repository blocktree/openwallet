/*
 * Copyright 2018 The OpenWallet Authors
 * This file is part of the OpenWallet library.
 *
 * The OpenWallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The OpenWallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */

package concurrent

//ProducerToConsumerRuntime 生产消费者运行模型
func ProducerToConsumerRuntime(producer chan interface{}, consumer chan interface{}) {

	var (
		values = make([]interface{}, 0)
	)

	for {

		var activeWorker chan<- interface{}
		var activeValue interface{}

		//当数据队列有数据时，释放顶部，传输给消费者
		if len(values) > 0 {
			activeWorker = consumer
			activeValue = values[0]
		}

		select {

		//生成者不断生成数据，插入到数据队列尾部
		case pa, exist := <-producer:
			if !exist {
				return
			}
			values = append(values, pa)
		case activeWorker <- activeValue:
			values = values[1:]
		}
	}

}
