// Package rand provides random utilities
package rand

import (
	"hash/maphash"
	"math/rand"
)

// This gives you a *rand.Rand which you can use as a source of randomness. It does not require locking - it uses a
// randomness seed in goroutine local storage, which is initialized by the runtime. It also doesn't suffer from the problem
// of using time.Now(), which is that the values are not actually very uniformly random. The returned values will be
// correlated, for similarly timed requests. The seeds from the runtime are instead derived (at least on linux) from a
// cryptographic entropy pool, so have a pretty good guarantee to be random. Lastly, it also gives you all the goodness
// from the rand package - for example, your modulo-logic introduces a bias towards small elements which (*rand.Rand).Intn
// doesn't have.
//
// @see: https://www.reddit.com/r/golang/comments/m9b0yp/fastest_way_to_pick_uniformly_from_a_slice_from/

type source struct{}

func (source) Int63() int64 {
	v := new(maphash.Hash).Sum64()
	return int64(v >> 1)
}

func (source) Seed(_ int64) {
	// ignored
}

func (source) Uint64() uint64 {
	return new(maphash.Hash).Sum64()
}

// Intn returns a non-negative pseudo-random number in [0,n)
func Intn(n int) int {
	return rand.New(source{}).Intn(n)
}
