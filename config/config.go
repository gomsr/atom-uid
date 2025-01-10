package config

import (
	"github.com/micro-services-roadmap/uid-generator-go/generator"
	"github.com/micro-services-roadmap/uid-generator-go/worker"
)

type Config struct {
	IdAssigner worker.Type    `mapstructure:"id_assigner" json:"id_assigner" yaml:"id_assigner"`
	Generator  generator.Type `mapstructure:"generator" json:"generator" yaml:"generator"`
	Delta      int64          `mapstructure:"delta" json:"delta" yaml:"delta"`             // (28 bits): 当前时间 -  "2016-05-20"的增量值, 单位: 秒
	Worker     int64          `mapstructure:"worker" json:"worker" yaml:"worker"`          // (22 bits): 机器 id, 最多可支持约 420w 次机器启动
	Sequence   int64          `mapstructure:"sequence" json:"sequence" yaml:"sequence"`    // (13 bits): 每秒下的并发序列, 13 bits 可支持每秒 8192 个并发.
	EpochStr   string         `mapstructure:"epoch_str" json:"epoch_str" yaml:"epoch_str"` // "2016-05-20"
}
