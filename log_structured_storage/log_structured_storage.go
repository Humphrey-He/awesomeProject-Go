package log_structured_storage

import "sync"

type Record struct {
	Key       string
	Value     []byte
	Tombstone bool
}

type Location struct {
	Segment int
	Index   int
	Deleted bool
}

type Segment struct {
	records []Record
}

func (s *Segment) append(rec Record) int {
	s.records = append(s.records, rec)
	return len(s.records) - 1
}

type LogStore struct {
	mu             sync.RWMutex
	segments       []*Segment
	index          map[string]Location
	maxSegmentSize int
}

func NewLogStore(maxSegmentSize int) *LogStore {
	if maxSegmentSize <= 0 {
		maxSegmentSize = 1024
	}
	return &LogStore{
		segments:       []*Segment{{}},
		index:          map[string]Location{},
		maxSegmentSize: maxSegmentSize,
	}
}

func (s *LogStore) Put(key string, value []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	seg := s.activeSegmentLocked()
	idx := seg.append(Record{Key: key, Value: value})
	s.index[key] = Location{Segment: len(s.segments) - 1, Index: idx}
}

func (s *LogStore) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	seg := s.activeSegmentLocked()
	idx := seg.append(Record{Key: key, Tombstone: true})
	s.index[key] = Location{Segment: len(s.segments) - 1, Index: idx, Deleted: true}
}

func (s *LogStore) Get(key string) ([]byte, bool) {
	s.mu.RLock()
	loc, ok := s.index[key]
	if !ok || loc.Deleted {
		s.mu.RUnlock()
		return nil, false
	}
	seg := s.segments[loc.Segment]
	rec := seg.records[loc.Index]
	s.mu.RUnlock()
	return rec.Value, true
}

func (s *LogStore) Compact() {
	s.mu.Lock()
	defer s.mu.Unlock()
	newSeg := &Segment{}
	newIndex := map[string]Location{}
	for key, loc := range s.index {
		if loc.Deleted {
			continue
		}
		rec := s.segments[loc.Segment].records[loc.Index]
		idx := newSeg.append(rec)
		newIndex[key] = Location{Segment: 0, Index: idx}
	}
	s.segments = []*Segment{newSeg}
	s.index = newIndex
}

func (s *LogStore) activeSegmentLocked() *Segment {
	seg := s.segments[len(s.segments)-1]
	if len(seg.records) >= s.maxSegmentSize {
		seg = &Segment{}
		s.segments = append(s.segments, seg)
	}
	return seg
}
