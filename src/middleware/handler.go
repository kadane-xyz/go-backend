package middleware

import "kadane.xyz/go-backend/v2/src/cache"

type Handler struct {
	CacheInstance *cache.Cache
}
