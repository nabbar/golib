/*
MIT License

Copyright (c) 2023 Nicolas JUHEL

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package kvitem_test

import (
	"errors"
	"strings"
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/nabbar/golib/database/kvdriver"
	"github.com/nabbar/golib/database/kvitem"
	"github.com/nabbar/golib/database/kvtypes"
)

// Test types
type TestUser struct {
	ID    string
	Name  string
	Email string
}

// Mock storage for testing
type mockStorage struct {
	data map[string]TestUser
	mu   sync.RWMutex
}

func newMockStorage() *mockStorage {
	return &mockStorage{
		data: make(map[string]TestUser),
	}
}

func (m *mockStorage) get(key string) (TestUser, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if val, ok := m.data[key]; ok {
		return val, nil
	}
	return TestUser{}, errors.New("not found")
}

func (m *mockStorage) set(key string, model TestUser) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[key] = model
	return nil
}

func (m *mockStorage) del(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.data, key)
	return nil
}

func (m *mockStorage) list() ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	keys := make([]string, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys, nil
}

func (m *mockStorage) search(prefix string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var keys []string
	for k := range m.data {
		if strings.HasPrefix(k, prefix) {
			keys = append(keys, k)
		}
	}
	return keys, nil
}

func (m *mockStorage) walk(fct kvtypes.FctWalk[string, TestUser]) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for k, v := range m.data {
		if !fct(k, v) {
			break
		}
	}
	return nil
}

// Helper to create a test driver
func createTestDriver(storage *mockStorage) kvtypes.KVDriver[string, TestUser] {
	// Create comparison functions
	compareEqual := func(a, b string) bool {
		return a == b
	}

	compareContains := func(ref, part string) bool {
		return strings.Contains(ref, part)
	}

	compareEmpty := func(s string) bool {
		return s == ""
	}

	compare := kvtypes.NewCompare[string](compareEqual, compareContains, compareEmpty)

	var newFunc kvdriver.FuncNew[string, TestUser]
	newFunc = func() kvtypes.KVDriver[string, TestUser] {
		return kvdriver.New[string, TestUser](
			compare,
			newFunc,
			storage.get,
			storage.set,
			storage.del,
			storage.list,
			storage.search,
			storage.walk,
		)
	}

	return kvdriver.New[string, TestUser](
		compare,
		newFunc,
		storage.get,
		storage.set,
		storage.del,
		storage.list,
		storage.search,
		storage.walk,
	)
}

var _ = Describe("KV Item", func() {
	var (
		storage *mockStorage
		driver  kvtypes.KVDriver[string, TestUser]
		item    kvtypes.KVItem[string, TestUser]
		key     string
	)

	BeforeEach(func() {
		storage = newMockStorage()
		driver = createTestDriver(storage)
		key = "user-1"
		item = kvitem.New[string, TestUser](driver, key)
	})

	Describe("New", func() {
		It("should create a new item instance", func() {
			Expect(item).ToNot(BeNil())
		})

		It("should set the correct key", func() {
			Expect(item.Key()).To(Equal(key))
		})

		It("should accept any driver", func() {
			newItem := kvitem.New[string, TestUser](driver, "test-key")
			Expect(newItem).ToNot(BeNil())
			Expect(newItem.Key()).To(Equal("test-key"))
		})
	})

	Describe("Key", func() {
		It("should return the item's key", func() {
			Expect(item.Key()).To(Equal("user-1"))
		})

		It("should return correct key for different items", func() {
			item2 := kvitem.New[string, TestUser](driver, "user-2")
			Expect(item2.Key()).To(Equal("user-2"))
		})
	})

	Describe("Set and Get", func() {
		It("should set and get a value", func() {
			user := TestUser{
				ID:    "user-1",
				Name:  "Alice",
				Email: "alice@example.com",
			}

			item.Set(user)
			retrieved := item.Get()

			Expect(retrieved.Name).To(Equal("Alice"))
			Expect(retrieved.Email).To(Equal("alice@example.com"))
		})

		It("should return zero value when not set", func() {
			retrieved := item.Get()
			Expect(retrieved.ID).To(BeEmpty())
			Expect(retrieved.Name).To(BeEmpty())
		})

		It("should override previous value", func() {
			user1 := TestUser{ID: "user-1", Name: "Alice"}
			item.Set(user1)

			user2 := TestUser{ID: "user-1", Name: "Bob"}
			item.Set(user2)

			retrieved := item.Get()
			Expect(retrieved.Name).To(Equal("Bob"))
		})
	})

	Describe("Load", func() {
		BeforeEach(func() {
			user := TestUser{
				ID:    "user-1",
				Name:  "Alice",
				Email: "alice@example.com",
			}
			storage.set("user-1", user)
		})

		It("should load data from driver", func() {
			err := item.Load()
			Expect(err).To(BeNil())

			retrieved := item.Get()
			Expect(retrieved.Name).To(Equal("Alice"))
			Expect(retrieved.Email).To(Equal("alice@example.com"))
		})

		It("should return error for non-existent key", func() {
			item2 := kvitem.New[string, TestUser](driver, "non-existent")
			err := item2.Load()
			Expect(err).ToNot(BeNil())
		})

		It("should update internal state on load", func() {
			err := item.Load()
			Expect(err).To(BeNil())

			// Modify storage
			user := TestUser{ID: "user-1", Name: "Alice Updated"}
			storage.set("user-1", user)

			// Load again
			err = item.Load()
			Expect(err).To(BeNil())

			retrieved := item.Get()
			Expect(retrieved.Name).To(Equal("Alice Updated"))
		})
	})

	Describe("Store", func() {
		It("should store data to driver", func() {
			user := TestUser{
				ID:    "user-1",
				Name:  "Alice",
				Email: "alice@example.com",
			}

			item.Set(user)
			err := item.Store(false)
			Expect(err).To(BeNil())

			// Verify in storage
			stored, err := storage.get("user-1")
			Expect(err).To(BeNil())
			Expect(stored.Name).To(Equal("Alice"))
		})

		It("should not store if no changes and force=false", func() {
			// Load existing data
			user := TestUser{ID: "user-1", Name: "Alice"}
			storage.set("user-1", user)
			item.Load()

			// Don't modify, just store
			err := item.Store(false)
			Expect(err).To(BeNil())

			// Should still be the same
			stored, _ := storage.get("user-1")
			Expect(stored.Name).To(Equal("Alice"))
		})

		It("should store even without changes when force=true", func() {
			user := TestUser{ID: "user-1", Name: "Alice"}
			storage.set("user-1", user)
			item.Load()

			// Force store without changes
			err := item.Store(true)
			Expect(err).To(BeNil())

			stored, _ := storage.get("user-1")
			Expect(stored.Name).To(Equal("Alice"))
		})

		It("should store modifications", func() {
			// Initial data
			user := TestUser{ID: "user-1", Name: "Alice"}
			storage.set("user-1", user)
			item.Load()

			// Modify
			user.Name = "Alice Modified"
			item.Set(user)

			// Store
			err := item.Store(false)
			Expect(err).To(BeNil())

			// Verify modification
			stored, _ := storage.get("user-1")
			Expect(stored.Name).To(Equal("Alice Modified"))
		})
	})

	Describe("Remove", func() {
		BeforeEach(func() {
			user := TestUser{ID: "user-1", Name: "Alice"}
			storage.set("user-1", user)
		})

		It("should remove item from storage", func() {
			err := item.Remove()
			Expect(err).To(BeNil())

			// Verify removal
			_, err = storage.get("user-1")
			Expect(err).ToNot(BeNil())
		})

		It("should not error when removing non-existent item", func() {
			item2 := kvitem.New[string, TestUser](driver, "non-existent")
			err := item2.Remove()
			Expect(err).To(BeNil())
		})

		It("should clear from storage permanently", func() {
			item.Remove()

			// Try to load after removal
			err := item.Load()
			Expect(err).ToNot(BeNil())
		})
	})

	Describe("Clean", func() {
		It("should clear internal state", func() {
			user := TestUser{ID: "user-1", Name: "Alice"}
			item.Set(user)

			item.Clean()

			retrieved := item.Get()
			Expect(retrieved.Name).To(BeEmpty())
			Expect(retrieved.Email).To(BeEmpty())
		})

		It("should reset both load and store models", func() {
			user := TestUser{ID: "user-1", Name: "Alice"}
			storage.set("user-1", user)
			item.Load()
			item.Set(user)

			item.Clean()

			retrieved := item.Get()
			Expect(retrieved.Name).To(BeEmpty())
		})

		It("should not affect storage", func() {
			user := TestUser{ID: "user-1", Name: "Alice"}
			storage.set("user-1", user)
			item.Load()

			item.Clean()

			// Storage should still have data
			stored, err := storage.get("user-1")
			Expect(err).To(BeNil())
			Expect(stored.Name).To(Equal("Alice"))
		})
	})

	Describe("HasChange", func() {
		It("should return true after load without set", func() {
			// After Load(), ml is set but ms is empty, so HasChange() returns true
			user := TestUser{ID: "user-1", Name: "Alice"}
			storage.set("user-1", user)
			item.Load()

			// Load() sets ml but not ms, so they differ
			Expect(item.HasChange()).To(BeTrue())
		})

		It("should return true when data is modified", func() {
			user := TestUser{ID: "user-1", Name: "Alice"}
			storage.set("user-1", user)
			item.Load()

			// Modify
			user.Name = "Alice Modified"
			item.Set(user)

			Expect(item.HasChange()).To(BeTrue())
		})

		It("should return true when new data is set without load", func() {
			user := TestUser{ID: "user-1", Name: "Alice"}
			item.Set(user)

			Expect(item.HasChange()).To(BeTrue())
		})

		It("should return false after storing changes", func() {
			user := TestUser{ID: "user-1", Name: "Alice"}
			item.Set(user)
			item.Store(true)

			// Load to sync state
			item.Load()

			Expect(item.HasChange()).To(BeFalse())
		})
	})

	Describe("Real-world scenarios", func() {
		It("should support full CRUD workflow", func() {
			// Create
			user := TestUser{
				ID:    "user-1",
				Name:  "Alice",
				Email: "alice@example.com",
			}
			item.Set(user)
			err := item.Store(false)
			Expect(err).To(BeNil())

			// Read
			item2 := kvitem.New[string, TestUser](driver, "user-1")
			err = item2.Load()
			Expect(err).To(BeNil())
			loaded := item2.Get()
			Expect(loaded.Name).To(Equal("Alice"))

			// Update
			loaded.Email = "alice.new@example.com"
			item2.Set(loaded)
			err = item2.Store(false)
			Expect(err).To(BeNil())

			// Verify update
			item3 := kvitem.New[string, TestUser](driver, "user-1")
			item3.Load()
			updated := item3.Get()
			Expect(updated.Email).To(Equal("alice.new@example.com"))

			// Delete
			err = item3.Remove()
			Expect(err).To(BeNil())

			// Verify deletion
			item4 := kvitem.New[string, TestUser](driver, "user-1")
			err = item4.Load()
			Expect(err).ToNot(BeNil())
		})

		It("should handle optimistic updates", func() {
			user := TestUser{ID: "user-1", Name: "Alice"}
			storage.set("user-1", user)

			// Load data
			item.Load()
			original := item.Get()
			Expect(original.Name).To(Equal("Alice"))

			// Simulate concurrent modification
			user.Name = "Alice Concurrent"
			storage.set("user-1", user)

			// Our modification
			modified := item.Get()
			modified.Email = "alice@example.com"
			item.Set(modified)

			// Store will overwrite concurrent changes
			item.Store(false)

			// Verify our changes won
			stored, _ := storage.get("user-1")
			Expect(stored.Email).To(Equal("alice@example.com"))
			Expect(stored.Name).To(Equal("Alice")) // Original name, not concurrent
		})

		It("should support conditional saves with HasChange", func() {
			user := TestUser{ID: "user-1", Name: "Alice"}
			storage.set("user-1", user)
			item.Load()

			// After Load(), HasChange() is true (ml set, ms empty)
			// So we set the loaded value to sync
			loaded := item.Get()
			item.Set(loaded)

			// Now both ml and ms have the same value
			initialHasChange := item.HasChange()
			Expect(initialHasChange).To(BeFalse())

			// Make changes
			loaded.Name = "Alice Modified"
			item.Set(loaded)

			// Conditional save
			if item.HasChange() {
				err := item.Store(false)
				Expect(err).To(BeNil())
			}

			// Verify save
			stored, _ := storage.get("user-1")
			Expect(stored.Name).To(Equal("Alice Modified"))
		})

		It("should handle load-modify-store pattern", func() {
			// Initial data
			user := TestUser{ID: "user-1", Name: "Alice", Email: "alice@example.com"}
			storage.set("user-1", user)

			// Load
			err := item.Load()
			Expect(err).To(BeNil())

			// Modify
			loaded := item.Get()
			loaded.Name = "Alice Updated"
			loaded.Email = "alice.updated@example.com"
			item.Set(loaded)

			// Verify change detection
			Expect(item.HasChange()).To(BeTrue())

			// Store
			err = item.Store(false)
			Expect(err).To(BeNil())

			// Verify persistence
			item2 := kvitem.New[string, TestUser](driver, "user-1")
			item2.Load()
			verified := item2.Get()
			Expect(verified.Name).To(Equal("Alice Updated"))
			Expect(verified.Email).To(Equal("alice.updated@example.com"))
		})

		It("should support clean and reuse pattern", func() {
			// First use
			user1 := TestUser{ID: "user-1", Name: "Alice"}
			item.Set(user1)
			item.Store(true)

			// Clean for reuse
			item.Clean()

			// Verify cleaned
			retrieved := item.Get()
			Expect(retrieved.Name).To(BeEmpty())

			// Reuse with new data
			user2 := TestUser{ID: "user-1", Name: "Bob"}
			item.Set(user2)
			item.Store(true)

			// Verify new data
			item.Load()
			final := item.Get()
			Expect(final.Name).To(Equal("Bob"))
		})
	})

	Describe("Edge cases", func() {
		It("should handle empty struct", func() {
			emptyUser := TestUser{}
			item.Set(emptyUser)

			retrieved := item.Get()
			Expect(retrieved.ID).To(BeEmpty())
			Expect(retrieved.Name).To(BeEmpty())
		})

		It("should handle multiple Set calls", func() {
			for i := 0; i < 10; i++ {
				user := TestUser{Name: "User " + string(rune('0'+i))}
				item.Set(user)
			}

			// Should have last value
			retrieved := item.Get()
			Expect(retrieved.Name).To(Equal("User 9"))
		})

		It("should maintain key integrity", func() {
			originalKey := item.Key()

			item.Set(TestUser{Name: "Alice"})
			item.Store(true)
			item.Load()
			item.Clean()

			Expect(item.Key()).To(Equal(originalKey))
		})
	})
})
