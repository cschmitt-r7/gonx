package gonx

import (
	"bufio"
	"io"
	"sync"
)

func (m *Map) handleError(err error) {
	//fmt.Fprintln(os.Stderr, err)
}

// Log Entry map
type Map struct {
	parser  *Parser
	entries chan Entry
	wg      sync.WaitGroup
}

func oneFileChannel(file io.Reader) chan io.Reader {
	ch := make(chan io.Reader, 1)
	ch <- file
	close(ch)
	return ch
}

func NewMap(files chan io.Reader, parser *Parser) *Map {
	m := &Map{
		parser:  parser,
		entries: make(chan Entry, 10),
	}

	for file := range files {
		m.wg.Add(1)
		go m.readFile(file)
	}

	go func() {
		m.wg.Wait()
		close(m.entries)
	}()

	return m
}

func (m *Map) readFile(file io.Reader) {
	// Iterate over log file lines and spawn new mapper goroutine
	// to parse it into given format
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		m.wg.Add(1)
		go func(line string) {
			entry, err := m.parser.ParseString(line)
			if err == nil {
				m.entries <- entry
			} else {
				m.handleError(err)
			}
			m.wg.Done()
		}(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		m.handleError(err)
	}
	m.wg.Done()
}

// Read next Entry from Entries channel. Return nil if channel is closed
func (m *Map) GetEntry() *Entry {
	entry, ok := <-m.entries
	if !ok {
		return nil
	}
	return &entry
}
