package cache

func init() {
	// enable cache for test runs
	SetCacheSize(32 * 1024 * 1024)
}
