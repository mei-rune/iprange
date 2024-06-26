package iprange

import (
	"net"
	"strings"
)

type numIterator struct {
	cur, end uint8
}

func (self *numIterator) Next() bool {
	if self.cur >= self.end {
		return false
	}
	self.cur++
	return true
}

func (self *numIterator) Current() uint8 {
	return self.cur
}

type ipSegments struct {
	s            string
	start1, end1 uint8
	start2, end2 uint8
	start3, end3 uint8
	start4, end4 uint8
}

func (self *ipSegments) String() string {
	return self.s
}

func (self *ipSegments) In(s string) bool {
	ip := net.ParseIP(s)
	if nil == ip {
		return false
	}
	
	return self.InAddr(ip)
}

func (self *ipSegments) InAddr(ip net.IP) bool {
	ip = ip.To4()
	if ip == nil {
		return false
	}

	if self.start1 > ip[0] || ip[0] > self.end1 {
		return false
	}

	if self.start2 > ip[1] || ip[1] > self.end2 {
		return false
	}

	if self.start3 > ip[2] || ip[2] > self.end3 {
		return false
	}

	if self.start4 > ip[3] || ip[3] > self.end4 {
		return false
	}
	return true
}

func (self *ipSegments) Iterable() Iterator {
	return &ipSegmentIterator{
		start1: self.start1,
		start2: self.start2,
		start3: self.start3,
		start4: self.start4,

		iter1: numIterator{cur: self.start1, end: self.end1},
		iter2: numIterator{cur: self.start2, end: self.end2},
		iter3: numIterator{cur: self.start3, end: self.end3},
		iter4: numIterator{cur: self.start4 - 1, end: self.end4},
	}
}

type ipSegmentIterator struct {
	start1 uint8
	start2 uint8
	start3 uint8
	start4 uint8

	iter1 numIterator
	iter2 numIterator
	iter3 numIterator
	iter4 numIterator
}

func (self *ipSegmentIterator) Next() bool {
	if self.iter4.Next() {
		return true
	}

	if self.iter3.Next() {
		self.iter4.cur = self.start4
		return true
	}

	if self.iter2.Next() {
		self.iter4.cur = self.start4
		self.iter3.cur = self.start3
		return true
	}

	if self.iter1.Next() {
		self.iter4.cur = self.start4
		self.iter3.cur = self.start3
		self.iter2.cur = self.start2
		return true
	}
	return false
}

func (self *ipSegmentIterator) Current() net.IP {
	return net.IPv4(
		self.iter1.Current(),
		self.iter2.Current(),
		self.iter3.Current(),
		self.iter4.Current(),
	)
}

type Ranges []Range

func (ranges Ranges) String() string {
	var sb strings.Builder
	for idx := range ranges {
		if idx > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(ranges[idx].String())
	}
	return sb.String()
}

func (ranges Ranges) In(s string) bool {
	ip := net.ParseIP(s)
	return ranges.InAddr(ip)
}

func (ranges Ranges) InAddr(ip net.IP) bool {
	for _, ra := range ranges {
		if ra.InAddr(ip) {
			return true
		}
	}
	return false
}

func (ranges Ranges) Iterable() Iterator {
	var inner = make([]Iterator, len(ranges))
	for idx := range ranges {
		inner[idx] = ranges[idx].Iterable()
	}
	return &Iterators{
		inner: inner,
	}
}

type Iterators struct {
	inner []Iterator
	index int
}

func (self *Iterators) Next() bool {
	for {
		if self.index >= len(self.inner) {
			return false
		}
		if self.inner[self.index].Next() {
			return true
		}
		self.index++
	}
}

func (self *Iterators) Current() net.IP {
	return self.inner[self.index].Current()
}
