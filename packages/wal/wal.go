package wal

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/prometheus/client_golang/prometheus"
)

type WAL struct {
	dir      string
	log      *logger.Logger
	metrics  *walMetrics
	segments map[uint32]*segment
	synced   map[uint32]bool
	mu       sync.RWMutex //nolint
}

type chainWAL struct {
	*WAL
	chainID *iscp.ChainID
}

func New(log *logger.Logger, dir string) *WAL {
	return &WAL{log: log, dir: dir, metrics: newWALMetrics(), synced: make(map[uint32]bool)}
}

var _ chain.WAL = &chainWAL{}

type segmentFile interface {
	Stat() (os.FileInfo, error)
	io.Writer
	io.Closer
	io.Reader
}

type segment struct {
	segmentFile
	index uint32
	dir   string
}

func (w *WAL) NewChainWAL(chainID *iscp.ChainID) (chain.WAL, error) {
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

	w.segments = make(map[uint32]*segment)
	files, _ := f.ReadDir(-1)
	for _, file := range files {
		w.metrics.segments.Inc()
		index, _ := strconv.ParseUint(file.Name(), 10, 32)
		w.segments[uint32(index)] = &segment{index: uint32(index), dir: w.dir}
	}
	return &chainWAL{w, chainID}, nil
}

func (w *chainWAL) Write(bytes []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	block, err := state.BlockFromBytes(bytes)
	if err != nil {
		return fmt.Errorf("Invalid block: %w", err)
	}
	segment, err := w.createSegment(block.BlockIndex())
	if err != nil {
		w.metrics.failedWrites.Inc()
		return fmt.Errorf("VMError writing log: %w", err)
	}
	n, err := segment.Write(bytes)
	if err != nil || len(bytes) != n {
		w.metrics.failedReads.Inc()
		return fmt.Errorf("VMError writing log: %w", err)
	}
	w.metrics.segments.Inc()
	return segment.Close()
}

func (w *chainWAL) createSegment(i uint32) (*segment, error) {
	segName := segmentName(w.dir, i)
	f, err := os.OpenFile(segName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o666)
	if err != nil {
		return nil, fmt.Errorf("could not create segment: %w", err)
	}
	s := &segment{index: i, segmentFile: f, dir: w.dir}
	w.segments[i] = s
	return s, nil
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
	if err := segment.load(); err != nil {
		w.metrics.failedReads.Inc()
		return nil, fmt.Errorf("VMError opening backup file: %w", err)
	}
	stat, err := segment.Stat()
	if err != nil {
		w.metrics.failedReads.Inc()
		return nil, fmt.Errorf("VMError reading backup file: %w", err)
	}
	blockBytes := make([]byte, stat.Size())
	bufr := bufio.NewReader(segment)
	n, err := bufr.Read(blockBytes)
	if err != nil || int64(n) != stat.Size() {
		w.metrics.failedReads.Inc()
		return nil, fmt.Errorf("VMError reading backup file: %w", err)
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

func (s *segment) load() error {
	segName := segmentName(s.dir, s.index)
	f, err := os.OpenFile(segName, os.O_RDONLY, 0o666)
	if err != nil {
		return fmt.Errorf("error opening segment: %w", err)
	}
	s.segmentFile = f
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
