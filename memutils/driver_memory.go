package memutils

import (
	"reflect"
	"strings"
	"sync"
	"time"
)

type MemDriverMemory struct {
	MemDriver

	mem   map[string]interface{}
	timer map[string]int64

	lock sync.Mutex
}

func (md *MemDriverMemory) Init(kargs ...string) {
	md.mem = make(map[string]interface{})
	md.timer = make(map[string]int64)
}

func (md *MemDriverMemory) unsafeRead(key string) (interface{}, bool) {
	now := Now()
	v, ok := md.mem[key]
	if ok {
		if t, ok := md.timer[key]; ok {
			if t <= now {
				delete(md.mem, key)
				delete(md.timer, key)
				return nil, false
			}
			return v, true
		} else {
			delete(md.mem, key)
			return nil, false
		}
	}
	return nil, false
}

func (md *MemDriverMemory) Read(key string) (interface{}, bool) {
	md.lock.Lock()
	defer md.lock.Unlock()

	return md.unsafeRead(key)
}

func (md *MemDriverMemory) unsafeWrite(key string, value interface{}, expire time.Duration, overwriteTTLIfExists bool) interface{} {
	now := Now()
	_, ok := md.mem[key]
	md.mem[key] = value
	if !ok || overwriteTTLIfExists {
		md.timer[key] = now + expire.Nanoseconds()
	}

	return value
}

func (md *MemDriverMemory) Write(key string, value interface{}, expire time.Duration, overwriteTTLIfExists bool) interface{} {
	md.lock.Lock()
	defer md.lock.Unlock()

	return md.unsafeWrite(key, value, expire, overwriteTTLIfExists)
}

func (md *MemDriverMemory) IncBy(key string, value int, expire time.Duration, overwriteTTLIfExists bool) int {
	md.lock.Lock()
	defer md.lock.Unlock()

	val, ok := md.unsafeRead(key)
	nextVal := value
	if ok && val != nil && reflect.TypeOf(val).Kind() == reflect.Int {
		nextVal = val.(int) + value
	}

	md.unsafeWrite(key, nextVal, expire, overwriteTTLIfExists)
	return nextVal
}

func (md *MemDriverMemory) Inc(key string, expire time.Duration, overwriteTTLIfExists bool) int {
	return md.IncBy(key, 1, expire, overwriteTTLIfExists)
}

func (md *MemDriverMemory) Exists(key string) bool {
	md.lock.Lock()
	defer md.lock.Unlock()

	now := Now()
	if t, ok := md.timer[key]; ok && t > now {
		if _, ok := md.mem[key]; ok {
			return true
		}
	}
	return false
}

func (md *MemDriverMemory) Expire(key string) {
	md.lock.Lock()
	defer md.lock.Unlock()

	_, ok := md.mem[key]
	if ok {
		delete(md.mem, key)
	}
	_, ok = md.timer[key]
	if ok {
		delete(md.timer, key)
	}
}

func (md *MemDriverMemory) SetExpire(key string, duration time.Duration) time.Duration {
	md.lock.Lock()
	defer md.lock.Unlock()

	if _, ok := md.unsafeRead(key); ok {
		md.timer[key] = Now() + duration.Nanoseconds()
	}
	return duration
}

func (md *MemDriverMemory) List(key string) []string {
	md.lock.Lock()
	defer md.lock.Unlock()

	slice := []string{}
	now := Now()
	for k, v := range md.timer {
		if now < v {
			slice = append(slice, k)
		}
	}
	return slice
}

func (md *MemDriverMemory) Wipe(prefix string) {
	md.lock.Lock()
	defer md.lock.Unlock()

	md.Init()
}

func (md *MemDriverMemory) WipePrefix(prefix string) {
	md.lock.Lock()
	defer md.lock.Unlock()

	for k := range md.mem {
		if strings.HasPrefix(k, prefix) {
			delete(md.mem, k)
			delete(md.timer, k)
		}
	}
}
