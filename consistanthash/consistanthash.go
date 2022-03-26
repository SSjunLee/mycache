package consistanthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func(data []byte) uint32
type Map struct {
	hash     Hash
	replicas int            //虚拟节点的倍数
	keys     []int          //哈希环
	hashMap  map[int]string //键是虚拟节点的哈希值，值是真实节点的名称。
}

func New(replicas int, hash Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     hash,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

func (m *Map) Add(keys ...string) {
	for _, k := range keys {
		for i := 0; i < m.replicas; i++ {
			//创建虚拟节点
			//虚拟节点的名称是：`strconv.Itoa(i) + key`
			hash := int(m.hash([]byte(strconv.Itoa(i) + k)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = k
		}
	}
	sort.Ints(m.keys)
}

//输入一个key返回最近的节点
func (m *Map) Get(key string) string {
	if key == "" {
		return ""
	}
	k := int(m.hash([]byte(key)))
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= k
	})
	//顺时针找到第一个匹配的虚拟节点的下标 `idx`，
	//从 m.keys 中获取到对应的哈希值。如果 `idx == len(m.keys)`，说明应选择 `m.keys[0]`
	return m.hashMap[m.keys[idx%len(m.keys)]]

}
