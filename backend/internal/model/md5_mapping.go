package model

import "strings"

// MD5MappingKey identifies the first stored AVIF object for an original-byte MD5
// within one runtime storage instance. The Redis adapter adds its own namespace
// prefix; callers should not pre-compose Redis keys by hand.
type MD5MappingKey struct {
	StorageKey string
	MD5Hash    string
}

// MD5Mapping is a cache-preheat/repair payload from a scoped original MD5 to
// the UID that owns the reusable physical object for that storage instance.
type MD5Mapping struct {
	Key MD5MappingKey
	UID string
}

func NewMD5MappingKey(storageKey string, md5Hash string) MD5MappingKey {
	return MD5MappingKey{
		StorageKey: strings.TrimSpace(storageKey),
		MD5Hash:    strings.ToLower(strings.TrimSpace(md5Hash)),
	}
}

func (k MD5MappingKey) CacheScope() string {
	return k.StorageKey + ":" + k.MD5Hash
}

func (k MD5MappingKey) MutexScope() string {
	return k.StorageKey + "\x00" + k.MD5Hash
}

func ParseMD5MappingCacheScope(raw string) (MD5MappingKey, bool) {
	storageKey, md5Hash, ok := strings.Cut(strings.TrimSpace(raw), ":")
	if !ok {
		return MD5MappingKey{}, false
	}
	key := NewMD5MappingKey(storageKey, md5Hash)
	if key.StorageKey == "" || key.MD5Hash == "" {
		return MD5MappingKey{}, false
	}
	return key, true
}
