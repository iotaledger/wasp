# Profiling

Profiling is disabled by default.

Running a node with `profiling.enabled = true` (in config.json or via command 
line) will spawn a pprof server running on `profiling.bindAddress` 
(`http://localhost:6060` by default).

By accessing `http://<profiling.bindAddress>/debug/pprof/` there are some 
profiles available, but the best way to visualize this data is using `go tool`
(requires `graphviz` installed).

Examples:

```shell
go tool pprof -http=:8080 http://localhost:6060/debug/pprof/<heap, profile?seconds=30, block, mutex>
```

```shell
wget -O trace.out http://localhost:6060/debug/pprof/trace?seconds=5
go tool trace trace.out
```

(for the above we need to manually download the trace profile from the pprof 
page, not sure if there is an automatic way) hint: use `?` to see keyboard 
shortcuts when on the trace page.

Upon node shutdown there will be some profile files available on `./profiles`:

- block.pprof
- cpu.pprof
- goroutine.pprof
- mem.pprof
- mutex.pprof
- thread.pprof
- trace.pprof
  
These profiles can be explored in the same way as the pprof server resources 
(using `go tool`).

While running cluster tests on Linux, working directory is `/tmp/wasp-cluster/`, thus the profiling data for each node is written in `profiles` subdirectory for each specific node,
e.g., `/tmp/wasp-cluster/wasp0/profiles/`.

## Notes

- Profiling has a negative impact on performance. Use when in need.
- Profiles may have no content if the node is not shutdown gracefully. (to use 
in tests, the test should fail via assertions, not via `-timeout` parameter)
