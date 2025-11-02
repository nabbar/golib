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

package kvdriver_test

import (
	"errors"
	"strings"
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/nabbar/golib/database/kvdriver"
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

var _ = Describe("KV Driver", func() {
	var (
		storage *mockStorage
		driver  kvtypes.KVDriver[string, TestUser]
	)

	BeforeEach(func() {
		storage = newMockStorage()
		driver = createTestDriver(storage)
	})

	Describe("New", func() {
		It("should create a new driver instance", func() {
			Expect(driver).ToNot(BeNil())
		})

		It("should create a new independent instance", func() {
			newDriver := driver.New()
			Expect(newDriver).ToNot(BeNil())
			Expect(newDriver).ToNot(BeIdenticalTo(driver))
		})
	})

	Describe("Set and Get", func() {
		It("should store and retrieve a value", func() {
			user := TestUser{
				ID:    "user-1",
				Name:  "Alice",
				Email: "alice@example.com",
			}

			err := driver.Set("user-1", user)
			Expect(err).To(BeNil())

			var retrieved TestUser
			err = driver.Get("user-1", &retrieved)
			Expect(err).To(BeNil())
			Expect(retrieved).To(Equal(user))
		})

		It("should update an existing value", func() {
			user1 := TestUser{ID: "user-1", Name: "Alice", Email: "alice@example.com"}
			err := driver.Set("user-1", user1)
			Expect(err).To(BeNil())

			user2 := TestUser{ID: "user-1", Name: "Alice Updated", Email: "alice.new@example.com"}
			err = driver.Set("user-1", user2)
			Expect(err).To(BeNil())

			var retrieved TestUser
			err = driver.Get("user-1", &retrieved)
			Expect(err).To(BeNil())
			Expect(retrieved.Name).To(Equal("Alice Updated"))
			Expect(retrieved.Email).To(Equal("alice.new@example.com"))
		})

		It("should return error for non-existent key", func() {
			var user TestUser
			err := driver.Get("non-existent", &user)
			Expect(err).ToNot(BeNil())
		})

		It("should handle empty string key", func() {
			user := TestUser{ID: "", Name: "Empty Key User"}
			err := driver.Set("", user)
			Expect(err).To(BeNil())

			var retrieved TestUser
			err = driver.Get("", &retrieved)
			Expect(err).To(BeNil())
			Expect(retrieved.Name).To(Equal("Empty Key User"))
		})
	})

	Describe("Del", func() {
		BeforeEach(func() {
			user := TestUser{ID: "user-1", Name: "Alice"}
			driver.Set("user-1", user)
		})

		It("should delete an existing key", func() {
			err := driver.Del("user-1")
			Expect(err).To(BeNil())

			var user TestUser
			err = driver.Get("user-1", &user)
			Expect(err).ToNot(BeNil())
		})

		It("should not error when deleting non-existent key", func() {
			err := driver.Del("non-existent")
			Expect(err).To(BeNil())
		})
	})

	Describe("List", func() {
		It("should return empty list when no items", func() {
			keys, err := driver.List()
			Expect(err).To(BeNil())
			Expect(keys).To(BeEmpty())
		})

		It("should list all keys", func() {
			users := []TestUser{
				{ID: "user-1", Name: "Alice"},
				{ID: "user-2", Name: "Bob"},
				{ID: "user-3", Name: "Charlie"},
			}

			for _, user := range users {
				driver.Set(user.ID, user)
			}

			keys, err := driver.List()
			Expect(err).To(BeNil())
			Expect(keys).To(HaveLen(3))
			Expect(keys).To(ContainElements("user-1", "user-2", "user-3"))
		})

		It("should reflect deletions", func() {
			driver.Set("user-1", TestUser{ID: "user-1"})
			driver.Set("user-2", TestUser{ID: "user-2"})
			driver.Set("user-3", TestUser{ID: "user-3"})

			driver.Del("user-2")

			keys, err := driver.List()
			Expect(err).To(BeNil())
			Expect(keys).To(HaveLen(2))
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

		It("should find keys with matching prefix", func() {
			keys, err := driver.Search("admin-")
			Expect(err).To(BeNil())
			Expect(keys).To(HaveLen(2))
			Expect(keys).To(ContainElements("admin-1", "admin-2"))
		})

		It("should return empty when no matches", func() {
			keys, err := driver.Search("nonexistent-")
			Expect(err).To(BeNil())
			Expect(keys).To(BeEmpty())
		})

		It("should find single match", func() {
			keys, err := driver.Search("guest-")
			Expect(err).To(BeNil())
			Expect(keys).To(HaveLen(1))
			Expect(keys).To(ContainElement("guest-1"))
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
			err := driver.Walk(func(key string, model TestUser) bool {
				count++
				Expect(key).ToNot(BeEmpty())
				Expect(model.Name).ToNot(BeEmpty())
				return true
			})

			Expect(err).To(BeNil())
			Expect(count).To(Equal(3))
		})

		It("should allow early termination", func() {
			count := 0
			err := driver.Walk(func(key string, model TestUser) bool {
				count++
				return count < 2 // Stop after 2 items
			})

			Expect(err).To(BeNil())
			Expect(count).To(Equal(2))
		})

		It("should handle empty storage", func() {
			emptyStorage := newMockStorage()
			emptyDriver := createTestDriver(emptyStorage)

			count := 0
			err := emptyDriver.Walk(func(key string, model TestUser) bool {
				count++
				return true
			})

			Expect(err).To(BeNil())
			Expect(count).To(Equal(0))
		})

		It("should collect all items", func() {
			var collected []TestUser
			err := driver.Walk(func(key string, model TestUser) bool {
				collected = append(collected, model)
				return true
			})

			Expect(err).To(BeNil())
			Expect(collected).To(HaveLen(3))
		})
	})

	Describe("Real-world scenarios", func() {
		It("should handle CRUD operations", func() {
			// Create
			user := TestUser{ID: "user-1", Name: "Alice", Email: "alice@example.com"}
			err := driver.Set("user-1", user)
			Expect(err).To(BeNil())

			// Read
			var retrieved TestUser
			err = driver.Get("user-1", &retrieved)
			Expect(err).To(BeNil())
			Expect(retrieved.Name).To(Equal("Alice"))

			// Update
			retrieved.Email = "alice.new@example.com"
			err = driver.Set("user-1", retrieved)
			Expect(err).To(BeNil())

			// Verify update
			var updated TestUser
			err = driver.Get("user-1", &updated)
			Expect(err).To(BeNil())
			Expect(updated.Email).To(Equal("alice.new@example.com"))

			// Delete
			err = driver.Del("user-1")
			Expect(err).To(BeNil())

			// Verify deletion
			var deleted TestUser
			err = driver.Get("user-1", &deleted)
			Expect(err).ToNot(BeNil())
		})

		It("should handle batch operations", func() {
			// Create multiple users
			for i := 1; i <= 10; i++ {
				user := TestUser{
					ID:    string(rune('0' + i)),
					Name:  "User " + string(rune('0'+i)),
					Email: "user" + string(rune('0'+i)) + "@example.com",
				}
				driver.Set(user.ID, user)
			}

			// List all
			keys, err := driver.List()
			Expect(err).To(BeNil())
			Expect(keys).To(HaveLen(10))

			// Collect items to update (avoid write during read lock)
			var updates []struct {
				key   string
				model TestUser
			}
			driver.Walk(func(key string, model TestUser) bool {
				model.Name = "Updated " + model.Name
				updates = append(updates, struct {
					key   string
					model TestUser
				}{key, model})
				return true
			})

			// Apply updates
			for _, u := range updates {
				driver.Set(u.key, u.model)
			}

			// Verify updates
			count := 0
			driver.Walk(func(key string, model TestUser) bool {
				if strings.HasPrefix(model.Name, "Updated") {
					count++
				}
				return true
			})
			Expect(count).To(Equal(10))
		})
	})
})
