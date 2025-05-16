package config

import (
	"github.com/gomsr/atom-uid/generator"
	"github.com/gomsr/atom-uid/worker"
)

type Config struct {
	IdAssigner worker.Type    `mapstructure:"id_assigner" json:"id_assigner" yaml:"id_assigner"`
	Generator  generator.Type `mapstructure:"generator" json:"generator" yaml:"generator"`
	TimeBits   int            `mapstructure:"time_bits" json:"time_bits" yaml:"time_bits"`       // (28 bits): 当前时间 -  "2016-05-20"的增量值, 单位: 秒
	WorkerBits int            `mapstructure:"worker_bits" json:"worker_bits" yaml:"worker_bits"` // (22 bits): 机器 id, 最多可支持约 420w 次机器启动
	SeqBits    int            `mapstructure:"seq_bits" json:"seq_bits" yaml:"seq_bits"`          // (13 bits): 每秒下的并发序列, 13 bits 可支持每秒 8192 个并发.
	EpochStr   string         `mapstructure:"epoch_str" json:"epoch_str" yaml:"epoch_str"`       // "2016-05-20"
}
