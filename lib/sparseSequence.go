package lib

import (
	"fmt"
	"strconv"
	"sync"
)

/* Old datastructure which is not used anymore.
It was created to be able to retrieve easily the first element
non appearing in a sequence of integers.
Basicaly, this structure partitions integers by bucket of "32" integers.
Each time a bucket is full it is remove.
When inserting an int "i", we first check if the bucket "i // 32" exists.
If yes, we turn on the i % 32 bit of the bucket. Otherwise we add a new bucket
and turn on the corresponding bit.
To get the first non present int, we can iterate on all buckets and for each
of them, using bittricks retrieve the leftmost 0
*/

/* consts. Each buckets holds 32 elements */
var size_bits = uint32(5)
var bucket_size = uint32(1 << size_bits)
var full_bucket = uint32(0xffffffff)
var lut = map[uint32]uint32{0x80000000: 32, 0x40000000: 31, 0x20000000: 30, 0x10000000: 29,
	0x8000000: 28, 0x4000000: 27, 0x2000000: 26, 0x1000000: 25,
	0x800000: 24, 0x400000: 23, 0x200000: 22, 0x100000: 21,
	0x80000: 20, 0x40000: 19, 0x20000: 18, 0x10000: 17,
	0x8000: 16, 0x4000: 15, 0x2000: 14, 0x1000: 13,
	0x800: 12, 0x400: 11, 0x200: 10, 0x100: 9,
	0x80: 8, 0x40: 7, 0x20: 6, 0x10: 5,
	0x8: 4, 0x4: 3, 0x2: 2, 0x1: 1,
	0: 0}

type SparseSequence struct {
	lock        *sync.RWMutex
	content     map[uint32]uint32
	max_dropped uint32
}

/* The goal of this datastructure is to split a sequence
into buckets of 32 elements
*/
func NewSparseSequence() SparseSequence {
	return SparseSequence{
		lock:        &sync.RWMutex{},
		content:     map[uint32]uint32{},
		max_dropped: 0}
}

func max(a uint32, b uint32) uint32 {
	if a > b {
		return a
	} else {
		return b
	}
}

func (seq *SparseSequence) Insert(id uint32) {
	bucket := uint32(id >> size_bits)
	e := uint32(id & (bucket_size - 1))

	new_element := uint32(1 << e)

	seq.lock.Lock()
	defer seq.lock.Unlock()

	if element, ok := seq.content[bucket]; ok {
		new_element = element | new_element
	}

	if new_element == full_bucket {
		delete(seq.content, bucket)
		seq.max_dropped = max(bucket, (seq.max_dropped+1)*bucket_size)
	} else {
		seq.content[bucket] = new_element
	}
}

func (seq *SparseSequence) GetMinNotPresent() uint32 {
	seq.lock.RLock()
	defer seq.lock.RUnlock()

	if len(seq.content) == 0 {
		return seq.max_dropped
	}
	out := uint32(0xffffffff)
	max_until_now := uint32(0)
	for offset, bucket := range seq.content {
		if bucket != full_bucket {
			rightmost_0 := uint32(^bucket & (bucket + 1))
			log := lut[rightmost_0] - 1
			index := offset*bucket_size + log
			if index < out {
				out = index
			}
		}
		if offset > max_until_now {
			max_until_now = offset * bucket_size
		}
	}
	if out == 0xffffffff {
		return max_until_now + bucket_size
	} else {
		return out
	}
}

func (seq *SparseSequence) Print() {
	seq.lock.RLock()
	defer seq.lock.RUnlock()

	for offset, bucket := range seq.content {
		fmt.Println(offset, strconv.FormatInt(int64(bucket), 2))
	}
}
