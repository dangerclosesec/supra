// internal/service/cache.go
// // Example usage:
// func ExampleUsage() {
// 	// Initialize cache service
// 	cacheService := NewCacheService(CacheConfig{
// 		TTL:         5 * time.Minute,
// 		CleanupFreq: 1 * time.Minute,
// 	})
// 	defer cacheService.Close()

// 	// Example struct
// 	type User struct {
// 		ID   string
// 		Name string
// 	}

// 	// Store a user
// 	user := User{ID: "123", Name: "John"}
// 	ctx := context.Background()
// 	err := cacheService.Set(ctx, "user:123", user)
// 	if err != nil {
// 		// Handle error
// 	}

// 	// Retrieve a user
// 	var cachedUser User
// 	err = cacheService.Get(ctx, "user:123", &cachedUser)
// 	if err != nil {
// 		// Handle error
// 	}

//		// Get or set with a fetch function
//		var result User
//		err = cacheService.GetOrSet(ctx, "user:123", &result, func() (interface{}, error) {
//			// This function is called only if the value is not in cache
//			return User{ID: "123", Name: "John"}, nil
//		})
//		if err != nil {
//			// Handle error
//		}
//	}
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dangerclosesec/supra/internal/cache"
	"github.com/dangerclosesec/supra/internal/domain"
)

// CacheService provides caching functionality with type safety and error handling
type CacheService struct {
	cache *cache.InMemoryCache
}

// CacheConfig holds configuration for the cache service
type CacheConfig struct {
	TTL         time.Duration
	CleanupFreq time.Duration
}

// NewCacheService creates a new cache service
func NewCacheService(config CacheConfig) *CacheService {
	cache := cache.NewInMemoryCache(config.TTL, config.CleanupFreq)

	// Start the cleanup routine
	ctx := context.Background()
	cache.StartCleanup(ctx)

	return &CacheService{
		cache: cache,
	}
}

// Set stores a value in the cache with type safety
func (s *CacheService) Set(ctx context.Context, key string, value interface{}) error {
	// Validate inputs
	if key == "" {
		return domain.ErrInvalidInput
	}

	// Store the value
	s.cache.Set(ctx, key, value)
	return nil
}

// CheckNonce checks if a nonce exists in the cache
func (s *CacheService) CheckNonce(ctx context.Context, nonce string) (bool, error) {
	// Validate inputs
	if nonce == "" {
		return false, domain.ErrInvalidInput
	}

	// Check if nonce exists in cache
	_, found := s.cache.Get(ctx, nonce)
	if !found {
		return false, nil
	}

	// Delete the nonce from cache
	s.cache.Delete(ctx, nonce)

	return found, nil
}

// Get retrieves a value from the cache with type conversion
func (s *CacheService) Get(ctx context.Context, key string, result interface{}) error {
	// Validate inputs
	if key == "" {
		return domain.ErrInvalidInput
	}

	// Get from cache
	value, found := s.cache.Get(ctx, key)
	if !found {
		return domain.ErrNotFound
	}

	// Handle type conversion
	switch v := value.(type) {
	case []byte:
		if err := json.Unmarshal(v, result); err != nil {
			return fmt.Errorf("unmarshaling cached value: %w", err)
		}
	default:
		// For direct type assignment
		if err := assignValue(value, result); err != nil {
			return fmt.Errorf("assigning cached value: %w", err)
		}
	}

	return nil
}

// GetOrSet retrieves a value from cache or sets it if not found
func (s *CacheService) GetOrSet(ctx context.Context, key string, result interface{}, fetchFunc func() (interface{}, error)) error {
	// Try to get from cache first
	err := s.Get(ctx, key, result)
	if err == nil {
		return nil
	}

	if err != domain.ErrNotFound {
		return fmt.Errorf("getting from cache: %w", err)
	}

	// Fetch new value
	value, err := fetchFunc()
	if err != nil {
		return fmt.Errorf("fetching value: %w", err)
	}

	// Store in cache
	if err := s.Set(ctx, key, value); err != nil {
		return fmt.Errorf("storing in cache: %w", err)
	}

	// Assign the fetched value to result
	if err := assignValue(value, result); err != nil {
		return fmt.Errorf("assigning fetched value: %w", err)
	}

	return nil
}

// Delete removes a value from the cache
func (s *CacheService) Delete(ctx context.Context, key string) error {
	if key == "" {
		return domain.ErrInvalidInput
	}

	s.cache.Delete(ctx, key)
	return nil
}

// Close stops the cleanup routine
func (s *CacheService) Close() {
	s.cache.StopCleanup()
}

// assignValue handles type conversion for different types
func assignValue(src interface{}, dst interface{}) error {
	// If the destination is a pointer to the same type as source
	if v, ok := dst.(*interface{}); ok {
		*v = src
		return nil
	}

	// Convert to JSON and back for complex types
	data, err := json.Marshal(src)
	if err != nil {
		return fmt.Errorf("marshaling value: %w", err)
	}

	if err := json.Unmarshal(data, dst); err != nil {
		return fmt.Errorf("unmarshaling value: %w", err)
	}

	return nil
}
