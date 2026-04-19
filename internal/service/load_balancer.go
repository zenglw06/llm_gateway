package service

import (
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/zenglw/llm_gateway/internal/llm"
)

// LoadBalancer 负载均衡器接口
type LoadBalancer interface {
	Name() string
	Select(services []llm.Service) llm.Service
}

// RoundRobinLoadBalancer 轮询负载均衡器
type RoundRobinLoadBalancer struct {
	index uint64
}

func NewRoundRobinLoadBalancer() *RoundRobinLoadBalancer {
	return &RoundRobinLoadBalancer{}
}

func (lb *RoundRobinLoadBalancer) Name() string {
	return "round_robin"
}

func (lb *RoundRobinLoadBalancer) Select(services []llm.Service) llm.Service {
	if len(services) == 0 {
		return nil
	}
	idx := atomic.AddUint64(&lb.index, 1) % uint64(len(services))
	return services[idx]
}

// RandomLoadBalancer 随机负载均衡器
type RandomLoadBalancer struct {
	r *rand.Rand
}

func NewRandomLoadBalancer() *RandomLoadBalancer {
	return &RandomLoadBalancer{
		r: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (lb *RandomLoadBalancer) Name() string {
	return "random"
}

func (lb *RandomLoadBalancer) Select(services []llm.Service) llm.Service {
	if len(services) == 0 {
		return nil
	}
	idx := lb.r.Intn(len(services))
	return services[idx]
}

// WeightedRoundRobinLoadBalancer 加权轮询负载均衡器
// 注意：需要Service接口扩展Weight()方法才能使用
type WeightedRoundRobinLoadBalancer struct {
	index uint64
}

func NewWeightedRoundRobinLoadBalancer() *WeightedRoundRobinLoadBalancer {
	return &WeightedRoundRobinLoadBalancer{}
}

func (lb *WeightedRoundRobinLoadBalancer) Name() string {
	return "weighted_round_robin"
}

func (lb *WeightedRoundRobinLoadBalancer) Select(services []llm.Service) llm.Service {
	if len(services) == 0 {
		return nil
	}
	// 简单实现，默认所有服务权重相同，退化为轮询
	idx := atomic.AddUint64(&lb.index, 1) % uint64(len(services))
	return services[idx]
}
