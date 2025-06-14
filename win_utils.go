/*
Copyright 2025 Milan Suk

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this db except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"math"
	"strings"
	"time"
)

func OsTicks() int64 {
	return time.Now().UnixMilli()
}

func OsIsTicksIn(start_ticks int64, delay_ms int) bool {
	return OsTicks() < (start_ticks + int64(delay_ms))
}

func OsTime() float64 {
	return float64(OsTicks()) / 1000
}

func OsTimeZone() int {
	_, zn := time.Now().Zone()
	return zn / 3600
}

const g_special = "ěščřžťďňĚŠČŘŽŤĎŇáíýéóúůÁÍÝÉÓÚŮÄäÖöÜüẞß"

func OsIsTextWord(ch rune) bool {
	return ((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || (ch == '_')) || strings.ContainsRune(g_special, ch)
}

func OsNextPowOf2(n int) int {
	k := 1
	for k < n {
		k = k << 1
	}
	return k
}

// Ternary operator
func OsTrn(question bool, ret_true int, ret_false int) int {
	if question {
		return ret_true
	}
	return ret_false
}

func OsTrnFloat(question bool, ret_true float64, ret_false float64) float64 {
	if question {
		return ret_true
	}
	return ret_false
}
func OsTrnString(question bool, ret_true string, ret_false string) string {
	if question {
		return ret_true
	}
	return ret_false
}

func OsTrnBool(question bool, ret_true bool, ret_false bool) bool {
	if question {
		return ret_true
	}
	return ret_false
}

func OsMax(x, y int) int {
	if x < y {
		return y
	}
	return x
}
func OsMin(x, y int) int {
	if x > y {
		return y
	}
	return x
}
func OsClamp(v, min, max int) int {
	return OsMin(OsMax(v, min), max)
}

func OsAbs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func OsMaxFloat(x, y float64) float64 {
	if x < y {
		return y
	}
	return x
}
func OsMinFloat(x, y float64) float64 {
	if x > y {
		return y
	}
	return x
}

func OsMaxFloat32(x, y float32) float32 {
	if x < y {
		return y
	}
	return x
}
func OsMinFloat32(x, y float32) float32 {
	if x > y {
		return y
	}
	return x
}

func OsClampFloat(v, min, max float64) float64 {
	return OsMinFloat(OsMaxFloat(v, min), max)
}

func OsFAbs(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}

func OsRoundDown(v float64) float64 {
	return float64(int(v))
}
func OsRoundUp(v float64) int {
	if v > OsRoundDown(v) {
		return int(v + 1)
	}
	return int(v)
}

func OsRoundHalf(v float64) int {
	return int(math.Round(v))
}
