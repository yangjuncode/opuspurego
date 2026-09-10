// Package opuspurego dynamically loads libopus through purego, without CGo
// in the binding. PCM is interleaved for stereo; frame sizes and decode return
// values are samples per channel. Codec objects must be closed before Library.
package opuspurego
