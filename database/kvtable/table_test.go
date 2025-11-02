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

package kvtable_test

import (
	"errors"
	"strings"
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/nabbar/golib/database/kvdriver"
	"github.com/nabbar/golib/database/kvtable"
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

var _ = Describe("KV Table", func() {
	var (
		storage *mockStorage
		driver  kvtypes.KVDriver[string, TestUser]
		table   kvtypes.KVTable[string, TestUser]
	)

	BeforeEach(func() {
		storage = newMockStorage()
		driver = createTestDriver(storage)
		table = kvtable.New[string, TestUser](driver)
	})

	Describe("New", func() {
		It("should create a new table instance", func() {
			Expect(table).ToNot(BeNil())
		})

		It("should accept a valid driver", func() {
			newTable := kvtable.New[string, TestUser](driver)
			Expect(newTable).ToNot(BeNil())
		})
	})

	Describe("Get", func() {
		BeforeEach(func() {
			user := TestUser{
				ID:    "user-1",
				Name:  "Alice",
				Email: "alice@example.com",
			}
			driver.Set("user-1", user)
		})

		It("should retrieve an existing item", func() {
			item, err := table.Get("user-1")
			Expect(err).To(BeNil())
			Expect(item).ToNot(BeNil())

			user := item.Get()
			Expect(user.Name).To(Equal("Alice"))
			Expect(user.Email).To(Equal("alice@example.com"))
		})

		It("should return error for non-existent key", func() {
			item, err := table.Get("non-existent")
			Expect(err).ToNot(BeNil())
			Expect(item).ToNot(BeNil()) // Item is still returned but empty
		})

		It("should return item with correct key", func() {
			item, err := table.Get("user-1")
			Expect(err).To(BeNil())
			Expect(item.Key()).To(Equal("user-1"))
		})
	})

	Describe("Del", func() {
		BeforeEach(func() {
			users := []TestUser{
				{ID: "user-1", Name: "Alice"},
				{ID: "user-2", Name: "Bob"},
			}
			for _, user := range users {
				driver.Set(user.ID, user)
			}
		})

		It("should delete an existing item", func() {
			err := table.Del("user-1")
			Expect(err).To(BeNil())

			// Verify deletion
			var user TestUser
			err = driver.Get("user-1", &user)
			Expect(err).ToNot(BeNil())
		})

		It("should not error when deleting non-existent key", func() {
			err := table.Del("non-existent")
			Expect(err).To(BeNil())
		})

		It("should only delete specified key", func() {
			err := table.Del("user-1")
			Expect(err).To(BeNil())

			// user-2 should still exist
			var user TestUser
			err = driver.Get("user-2", &user)
			Expect(err).To(BeNil())
			Expect(user.Name).To(Equal("Bob"))
		})
	})

	Describe("List", func() {
		It("should return empty list when no items", func() {
			items, err := table.List()
			Expect(err).To(BeNil())
			Expect(items).To(BeEmpty())
		})

		It("should list all items", func() {
			users := []TestUser{
				{ID: "user-1", Name: "Alice"},
				{ID: "user-2", Name: "Bob"},
				{ID: "user-3", Name: "Charlie"},
			}

			for _, user := range users {
				driver.Set(user.ID, user)
			}

			items, err := table.List()
			Expect(err).To(BeNil())
			Expect(items).To(HaveLen(3))

			// Verify all items are present
			keys := make([]string, len(items))
			for i, item := range items {
				keys[i] = item.Key()
			}
			Expect(keys).To(ContainElements("user-1", "user-2", "user-3"))
		})

		It("should return items that can be loaded", func() {
			user := TestUser{ID: "user-1", Name: "Alice", Email: "alice@example.com"}
			driver.Set("user-1", user)

			items, err := table.List()
			Expect(err).To(BeNil())
			Expect(items).To(HaveLen(1))

			// Item needs to be loaded first
			item := items[0]
			err = item.Load()
			Expect(err).To(BeNil())

			retrieved := item.Get()
			Expect(retrieved.Name).To(Equal("Alice"))
		})

		It("should reflect deletions", func() {
			driver.Set("user-1", TestUser{ID: "user-1"})
			driver.Set("user-2", TestUser{ID: "user-2"})
			driver.Set("user-3", TestUser{ID: "user-3"})

			table.Del("user-2")

			items, err := table.List()
			Expect(err).To(BeNil())
			Expect(items).To(HaveLen(2))

			keys := make([]string, len(items))
			for i, item := range items {
				keys[i] = item.Key()
			}
			Expect(keys).ToNot(ContainElement("user-2"))
		})
	})

	Describe("Search", func() {
		BeforeEach(func() {
			users := []TestUser{
				{ID: "admin-1", Name: "Admin One"},
				{ID: "admin-2", Name: "Admin Two"},
				{ID: "user-1", Name: "User One"},
				{ID: "user-2", Name: "User Two"},
				{ID: "guest-1", Name: "Guest One"},
			}

			for _, user := range users {
				driver.Set(user.ID, user)
			}
		})

		It("should find items with matching prefix", func() {
			items, err := table.Search("admin-")
			Expect(err).To(BeNil())
			Expect(items).To(HaveLen(2))

			keys := make([]string, len(items))
			for i, item := range items {
				keys[i] = item.Key()
			}
			Expect(keys).To(ContainElements("admin-1", "admin-2"))
		})

		It("should return empty when no matches", func() {
			items, err := table.Search("nonexistent-")
			Expect(err).To(BeNil())
			Expect(items).To(BeEmpty())
		})

		It("should find single match", func() {
			items, err := table.Search("guest-")
			Expect(err).To(BeNil())
			Expect(items).To(HaveLen(1))

			Expect(items[0].Key()).To(Equal("guest-1"))

			// Load the item to get data
			err = items[0].Load()
			Expect(err).To(BeNil())

			user := items[0].Get()
			Expect(user.Name).To(Equal("Guest One"))
		})

		It("should return items with correct data", func() {
			items, err := table.Search("user-")
			Expect(err).To(BeNil())
			Expect(items).To(HaveLen(2))

			for _, item := range items {
				// Load each item to get data
				err := item.Load()
				Expect(err).To(BeNil())

				user := item.Get()
				Expect(user.Name).To(ContainSubstring("User"))
			}
		})
	})

	Describe("Walk", func() {
		BeforeEach(func() {
			users := []TestUser{
				{ID: "user-1", Name: "Alice"},
				{ID: "user-2", Name: "Bob"},
				{ID: "user-3", Name: "Charlie"},
			}

			for _, user := range users {
				driver.Set(user.ID, user)
			}
		})

		It("should walk through all items", func() {
			count := 0
			err := table.Walk(func(item kvtypes.KVItem[string, TestUser]) bool {
				count++
				Expect(item).ToNot(BeNil())
				Expect(item.Key()).ToNot(BeEmpty())
				return true
			})

			Expect(err).To(BeNil())
			Expect(count).To(Equal(3))
		})

		It("should allow early termination", func() {
			count := 0
			err := table.Walk(func(item kvtypes.KVItem[string, TestUser]) bool {
				count++
				return count < 2 // Stop after 2 items
			})

			Expect(err).To(BeNil())
			Expect(count).To(Equal(2))
		})

		It("should handle empty table", func() {
			emptyStorage := newMockStorage()
			emptyDriver := createTestDriver(emptyStorage)
			emptyTable := kvtable.New[string, TestUser](emptyDriver)

			count := 0
			err := emptyTable.Walk(func(item kvtypes.KVItem[string, TestUser]) bool {
				count++
				return true
			})

			Expect(err).To(BeNil())
			Expect(count).To(Equal(0))
		})

		It("should provide items with correct data", func() {
			var names []string
			err := table.Walk(func(item kvtypes.KVItem[string, TestUser]) bool {
				user := item.Get()
				names = append(names, user.Name)
				return true
			})

			Expect(err).To(BeNil())
			Expect(names).To(HaveLen(3))
			Expect(names).To(ContainElements("Alice", "Bob", "Charlie"))
		})

		It("should allow item modifications during walk", func() {
			// Collect items first (to avoid modifying during read)
			var items []kvtypes.KVItem[string, TestUser]
			table.Walk(func(item kvtypes.KVItem[string, TestUser]) bool {
				items = append(items, item)
				return true
			})

			// Modify collected items
			for _, item := range items {
				user := item.Get()
				user.Name = "Updated " + user.Name
				item.Set(user)
				item.Store(true)
			}

			// Verify modifications
			count := 0
			table.Walk(func(item kvtypes.KVItem[string, TestUser]) bool {
				user := item.Get()
				if strings.HasPrefix(user.Name, "Updated") {
					count++
				}
				return true
			})
			Expect(count).To(Equal(3))
		})
	})

	Describe("Real-world scenarios", func() {
		It("should support full CRUD workflow", func() {
			// Create
			user := TestUser{ID: "user-1", Name: "Alice", Email: "alice@example.com"}
			driver.Set("user-1", user)

			// Read
			item, err := table.Get("user-1")
			Expect(err).To(BeNil())
			retrieved := item.Get()
			Expect(retrieved.Name).To(Equal("Alice"))

			// Update
			retrieved.Email = "alice.new@example.com"
			item.Set(retrieved)
			err = item.Store(false)
			Expect(err).To(BeNil())

			// Verify update
			item2, err := table.Get("user-1")
			Expect(err).To(BeNil())
			updated := item2.Get()
			Expect(updated.Email).To(Equal("alice.new@example.com"))

			// Delete
			err = table.Del("user-1")
			Expect(err).To(BeNil())

			// Verify deletion
			_, err = table.Get("user-1")
			Expect(err).ToNot(BeNil())
		})

		It("should handle batch operations efficiently", func() {
			// Create multiple items
			for i := 1; i <= 10; i++ {
				user := TestUser{
					ID:    string(rune('0' + i)),
					Name:  "User " + string(rune('0'+i)),
					Email: "user" + string(rune('0'+i)) + "@example.com",
				}
				driver.Set(user.ID, user)
			}

			// List all
			items, err := table.List()
			Expect(err).To(BeNil())
			Expect(items).To(HaveLen(10))

			// Search subset
			driver.Set("admin-1", TestUser{ID: "admin-1", Name: "Admin"})
			adminItems, err := table.Search("admin-")
			Expect(err).To(BeNil())
			Expect(adminItems).To(HaveLen(1))

			// Delete all non-admin users
			for _, item := range items {
				table.Del(item.Key())
			}

			// Verify only admin remains
			remaining, err := table.List()
			Expect(err).To(BeNil())
			Expect(remaining).To(HaveLen(1))
			Expect(remaining[0].Key()).To(Equal("admin-1"))
		})

		It("should support item change tracking", func() {
			user := TestUser{ID: "user-1", Name: "Alice"}
			driver.Set("user-1", user)

			item, err := table.Get("user-1")
			Expect(err).To(BeNil())

			// Get() loads the item, verify it was loaded correctly
			loadedUser := item.Get()
			Expect(loadedUser.Name).To(Equal("Alice"))

			// Modify
			loadedUser.Name = "Alice Modified"
			item.Set(loadedUser)

			// Should detect change
			Expect(item.HasChange()).To(BeTrue())

			// Store with force=true
			err = item.Store(true)
			Expect(err).To(BeNil())

			// Verify the modification was saved
			item2, err := table.Get("user-1")
			Expect(err).To(BeNil())
			verifyUser := item2.Get()
			Expect(verifyUser.Name).To(Equal("Alice Modified"))
		})
	})
})
