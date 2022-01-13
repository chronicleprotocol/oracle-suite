package memory

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/chronicleprotocol/oracle-suite/pkg/transport/messages"
)

func TestMemory_Add(t *testing.T) {
	m := New(time.Minute)
	e1 := &messages.Event{
		Date:       time.Now(),
		Type:       "test",
		ID:         []byte("test"),
		Index:      []byte("idx"),
		Data:       map[string][]byte{"test": []byte("test")},
		Signatures: map[string][]byte{"test": []byte("test")},
	}
	e2 := &messages.Event{
		Date:       time.Now(),
		Type:       "test",
		ID:         []byte("test2"),
		Index:      []byte("idx"),
		Data:       map[string][]byte{"test": []byte("test2")},
		Signatures: map[string][]byte{"test": []byte("test2")},
	}
	e3 := &messages.Event{
		Date:       time.Now(),
		Type:       "test",
		ID:         []byte("test2"),
		Index:      []byte("idx2"),
		Data:       map[string][]byte{"test": []byte("test2")},
		Signatures: map[string][]byte{"test": []byte("test2")},
	}

	assert.NoError(t, m.Add(e1))
	assert.NoError(t, m.Add(e2))
	assert.NoError(t, m.Add(e3)) // different index

	es, err := m.Get("test", []byte("idx"))
	assert.NoError(t, err)
	assert.ElementsMatch(t, []*messages.Event{e1, e2}, es)
}

func TestMemory_gc(t *testing.T) {
	m := New(time.Minute)
	assert.NoError(t, m.Add(&messages.Event{
		Date:       time.Now(),
		Type:       "test",
		ID:         []byte("test"),
		Index:      []byte("idx"),
		Data:       map[string][]byte{"test": []byte("test")},
		Signatures: map[string][]byte{"test": []byte("test")},
	}))
	for i := 0; i < m.gcevery-1; i++ {
		e := &messages.Event{
			Date:       time.Unix(0, 0),
			Type:       "test",
			ID:         []byte(strconv.Itoa(i)),
			Index:      []byte("idx"),
			Data:       map[string][]byte{"test": []byte("test")},
			Signatures: map[string][]byte{"test": []byte("test")},
		}
		assert.NoError(t, m.Add(e))
	}

	es, err := m.Get("test", []byte("idx"))
	assert.NoError(t, err)
	assert.Len(t, es, 1)
}

func TestMemory_gc_allExpired(t *testing.T) {
	m := New(time.Minute)
	for i := 0; i < m.gcevery; i++ {
		e := &messages.Event{
			Date:       time.Unix(0, 0),
			Type:       "test",
			ID:         []byte(strconv.Itoa(i)),
			Index:      []byte("idx"),
			Data:       map[string][]byte{"test": []byte("test")},
			Signatures: map[string][]byte{"test": []byte("test")},
		}
		assert.NoError(t, m.Add(e))
	}

	es, err := m.Get("test", []byte("idx"))
	assert.NoError(t, err)
	assert.Len(t, es, 0)
}
