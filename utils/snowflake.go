/**
  @author: Bruce
  @since: 2023/5/19
  @desc: //TODO
**/

package utils

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

const (
	epoch             = int64(1577808000000)                           // 设置起始时间(时间戳/毫秒)：2020-01-01 00:00:00，有效期69年
	timestampBits     = uint(41)                                       // 时间戳占用位数
	datacenterIdBits  = uint(5)                                        // 数据中心id所占位数
	workerIdBits      = uint(5)                                        // 机器id所占位数
	sequenceBits      = uint(12)                                       // 序列所占的位数
	timestampMax      = int64(-1 ^ (-1 << timestampBits))              // 时间戳最大值
	datacenterIdMax   = int64(-1 ^ (-1 << datacenterIdBits))           // 支持的最大数据中心id数量
	workerIdMax       = int64(-1 ^ (-1 << workerIdBits))               // 支持的最大机器id数量
	sequenceMask      = int64(-1 ^ (-1 << sequenceBits))               // 支持的最大序列id数量
	workerIdShift     = sequenceBits                                   // 机器id左移位数
	datacenterIdShift = sequenceBits + workerIdBits                    // 数据中心id左移位数
	timestampShift    = sequenceBits + workerIdBits + datacenterIdBits // 时间戳左移位数
)

type SnowflakeGenerator struct {
	sync.Mutex
	timestamp    int64 // 时间戳 ，毫秒
	workerId     int64 // 工作节点
	datacenterId int64 // 数据中心机房id
	sequence     int64 // 序列号
}

func NewSnowflakeGenerator(workerId, datacenterId int64) *SnowflakeGenerator {
	return &SnowflakeGenerator{
		timestamp:    0,
		workerId:     workerId,
		datacenterId: datacenterId,
		sequence:     0,
	}
}

func (s *SnowflakeGenerator) NextVal() (uint64, error) {
	s.Lock()
	now := time.Now().UnixNano() / 1000000 // 转毫秒
	if s.timestamp == now {
		// 当同一时间戳（精度：毫秒）下多次生成id会增加序列号
		s.sequence = (s.sequence + 1) & sequenceMask
		if s.sequence == 0 {
			// 如果当前序列超出12bit长度，则需要等待下一毫秒
			// 下一毫秒将使用sequence:0
			for now <= s.timestamp {
				now = time.Now().UnixNano() / 1000000
			}
		}
	} else {
		// 不同时间戳（精度：毫秒）下直接使用序列号：0
		s.sequence = 0
	}
	t := now - epoch
	if t > timestampMax {
		s.Unlock()

		return 0, errors.New(fmt.Sprintf("epoch must be between 0 and %d", timestampMax-1))
	}
	s.timestamp = now
	r := uint64((t)<<timestampShift | (s.datacenterId << datacenterIdShift) | (s.workerId << workerIdShift) | (s.sequence))
	s.Unlock()
	return r, nil
}
