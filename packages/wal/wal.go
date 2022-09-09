package wal

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/prometheus/client_golang/prometheus"
)

type WAL struct {
	dir      string
	log      *logger.Logger
	metrics  *walMetrics
	segments map[uint32]*segment
	synced   map[uint32]bool
}

type chainWAL struct {
	*WAL
	chainID *isc.ChainID
	mu      sync.RWMutex
}

func New(log *logger.Logger, dir string) *WAL {
	return &WAL{log: log, dir: dir, metrics: newWALMetrics(), synced: make(map[uint32]bool)}
}

var _ chain.WAL = &chainWAL{}

type segment struct {
	index uint32
	dir   string
}

func (w *WAL) NewChainWAL(chainID *isc.ChainID) (chain.WAL, error) {
	if w == nil {
		return &defaultWAL{}, nil
	}
	w.dir = filepath.Join(w.dir, chainID.String())
	if err := os.MkdirAll(w.dir, 0o777); err != nil {
		return nil, fmt.Errorf("create dir: %w", err)
	}
	// read all segments in log
	f, err := os.Open(w.dir)
	if err != nil {
		return nil, fmt.Errorf("could not open wal: %w", err)
	}
	defer f.Close()

	w.segments = make(map[uint32]*segment)
	files, _ := f.ReadDir(-1)
	for _, file := range files {
		w.metrics.segments.Inc()
		index, _ := strconv.ParseUint(file.Name(), 10, 32)
		w.segments[uint32(index)] = &segment{index: uint32(index), dir: w.dir}
	}
	return &chainWAL{WAL: w, chainID: chainID}, nil
}

func (w *chainWAL) Write(bytes []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	block, err := state.BlockFromBytes(bytes)
	if err != nil {
		return fmt.Errorf("Invalid block: %w", err)
	}

	index := block.BlockIndex()
	segName := segmentName(w.dir, index)
	f, err := os.OpenFile(segName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o666)
	if err != nil {
		return fmt.Errorf("could not create segment: %w", err)
	}
	defer f.Close()
	segment := &segment{index: index, dir: w.dir}
	w.segments[index] = segment
	if err != nil {
		w.metrics.failedWrites.Inc()
		return fmt.Errorf("Error writing log: %w", err)
	}
	n, err := f.Write(bytes)
	if err != nil || len(bytes) != n {
		w.metrics.failedReads.Inc()
		return fmt.Errorf("Error writing log: %w", err)
	}
	w.metrics.segments.Inc()
	return nil
}

func segmentName(dir string, index uint32) string {
	return filepath.Join(dir, fmt.Sprintf("%010d", index))
}

func (w *chainWAL) Contains(i uint32) bool {
	return w.getSegment(i) != nil
}

func (w *chainWAL) Read(i uint32) ([]byte, error) {
	segment := w.getSegment(i)
	if segment == nil {
		return nil, fmt.Errorf("block not found in wal")
	}
	segName := segmentName(segment.dir, segment.index)
	f, err := os.OpenFile(segName, os.O_RDONLY, 0o666)
	if err != nil {
		w.metrics.failedReads.Inc()
		return nil, fmt.Errorf("error opening segment: %w", err)
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		w.metrics.failedReads.Inc()
		return nil, fmt.Errorf("Error reading backup file: %w", err)
	}
	blockBytes := make([]byte, stat.Size())
	bufr := bufio.NewReader(f)
	n, err := bufr.Read(blockBytes)
	if err != nil || int64(n) != stat.Size() {
		w.metrics.failedReads.Inc()
		return nil, fmt.Errorf("Error reading backup file: %w", err)
	}
	return blockBytes, nil
}

func (w *chainWAL) getSegment(i uint32) *segment {
	segment, ok := w.segments[i]
	if ok {
		return segment
	}
	return nil
}

type defaultWAL struct{}

var _ chain.WAL = &defaultWAL{}

func (w *defaultWAL) Write(_ []byte) error {
	return nil
}

func (w *defaultWAL) Read(i uint32) ([]byte, error) {
	return nil, fmt.Errorf("Empty wal")
}

func (w *defaultWAL) Contains(i uint32) bool {
	return false
}

func NewDefault() chain.WAL {
	return &defaultWAL{}
}

type walMetrics struct {
	segments     prometheus.Counter
	failedWrites prometheus.Counter
	failedReads  prometheus.Counter
}

var once sync.Once

func newWALMetrics() *walMetrics {
	m := &walMetrics{}

	m.segments = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "wasp_wal_total_segments",
		Help: "Total number of segment files",
	})

	m.failedWrites = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "wasp_wal_failed_writes",
		Help: "Total number of writes to WAL that failed",
	})

	m.failedReads = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "wasp_wal_failed_reads",
		Help: "Total number of reads failed while replaying WAL",
	})

	registerMetrics := func() {
		prometheus.MustRegister(
			m.segments,
			m.failedWrites,
			m.failedReads,
		)
	}
	once.Do(registerMetrics)
	return m
}
