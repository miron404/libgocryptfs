package main

// Strings received by the exported gcf_* functions are passed in by cgo as
// (pointer, length) pairs. The pointer belongs to the caller: DroidFS builds
// them from JNI GetStringUTFChars(), so the bytes live in a malloc()ed buffer,
// not on the Go heap.
//
// Go's assembly string primitives (internal/bytealg.IndexByteString,
// runtime.findnull, ...) are written to be *page*-safe, not *allocation*-safe:
// they read 16/32-byte blocks and deliberately run past the end of the data as
// long as they stay within the memory page. That is harmless on the Go heap,
// but the C heap is tagged when the process runs with ARM MTE (enabled by
// default for every app on GrapheneOS / ARMv9), and MTE granules are 16 bytes
// wide. The first block read past the end hits a granule carrying a different
// tag and the process dies with SIGSEGV / SEGV_MTESERR.
//
// Upstream bug: https://github.com/golang/go/issues/27610 (open, unplanned).
//
// fromCString returns a Go-owned copy so that no code downstream ever reads C
// memory. copy() compiles down to runtime.memmove, which reads exactly len(s)
// bytes and is therefore safe on tagged memory. This also removes the
// lifetime hazard of retaining substrings (filepath.Base/Dir return slices of
// their input) past the caller's ReleaseStringUTFChars().
func fromCString(s string) string {
	b := make([]byte, len(s))
	copy(b, s)
	return string(b)
}
