package cache

func init() {
	// enable cache for test runs
	InitCache(32 * 1024 * 1024)
}
