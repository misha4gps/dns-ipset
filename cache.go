package main

import (
	"github.com/miekg/dns"
	"sync"
	"time"
)

type Cache interface {
	Get(reqType uint16, domain string) dns.RR
	Set(reqType uint16, domain string, ip dns.RR)
}

type CacheItem struct {
	Ip  []dns.RR
	Die time.Time
}

type MemoryCache struct {
	cache  map[uint16]map[string]*CacheItem
	locker sync.RWMutex
}

func NewMemoryCache() *MemoryCache {
	cache := &MemoryCache{
		cache: make(map[uint16]map[string]*CacheItem),
	}
	go cache.cleaner()

	return cache
}

func (c *MemoryCache) Get(reqType uint16, domain string) []dns.RR {
	c.locker.RLock()
	defer c.locker.RUnlock()

	if m, ok := c.cache[reqType]; ok {
		if ip, ok := m[domain]; ok {
			if ip.Die.After(time.Now()) {
				for _, ipV := range ip.Ip {
					ipV.Header().Ttl = uint32(ip.Die.Sub(time.Now()) / time.Second)
				}
				return ip.Ip
			}
		}
	}

	return nil
}

func (c *MemoryCache) Set(reqType uint16, domain string, answers []dns.RR) {
	c.locker.Lock()
	defer c.locker.Unlock()

	var m map[string]*CacheItem

	m, ok := c.cache[reqType]
	if !ok {
		m = make(map[string]*CacheItem)
		c.cache[reqType] = m
	}

	m[domain] = &CacheItem{
		Ip: answers,
	}

	if len(answers) == 0 {
		m[domain].Die = time.Now().Add(10 * time.Second)
		return
	}

	minTtl := answers[0].Header().Ttl
	for _, answer := range answers {
		if answer.Header().Ttl < minTtl {
			minTtl = answer.Header().Ttl
		}
	}
	if minTtl > 1800 {
		minTtl = 1800
	}
	m[domain].Die = time.Now().Add(time.Duration(minTtl) * time.Second)
}

func (c *MemoryCache) cleaner() {
	for c != nil {
		c.locker.Lock()
		now := time.Now()

		for _, v := range c.cache {
			for k, vv := range v {
				if vv.Die.Before(now) {
					delete(v, k)
				}
			}
		}

		c.locker.Unlock()
		time.Sleep(config.UpdateInterval)
	}
}
