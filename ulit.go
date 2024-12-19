/*
Copyright 2024 Milan Suk

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
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strconv"
	"strings"
	"time"

	"github.com/pbnjay/memory"
)

func OsFormatBytes(bytes int) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div := int64(unit)
	exp := 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
func OsFormatBytes2(bytesA, bytesB int) string {
	units := []string{"B", "KB", "MB", "GB", "TB", "PB", "EB"}
	const unit = 1024

	getUnit := func(bytes int) int {
		exp := 0
		for bytes >= unit && exp < len(units)-1 {
			bytes /= unit
			exp++
		}
		return exp
	}

	unitA, unitB := getUnit(bytesA), getUnit(bytesB)
	lowerUnit := unitA
	if unitB < unitA {
		lowerUnit = unitB
	}

	divider := math.Pow(float64(unit), float64(lowerUnit))
	valueA := float64(bytesA) / divider
	valueB := float64(bytesB) / divider

	if lowerUnit == 0 {
		return fmt.Sprintf("%.0f/%.0f %s", valueA, valueB, units[lowerUnit])
	}
	return fmt.Sprintf("%.1f/%.1f %s", valueA, valueB, units[lowerUnit])
}

func OsMarshal(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		log.Fatal("OsMarshal failed:", err.Error())
	}
	return data

	//b := new(bytes.Buffer)
	//err := gob.NewEncoder(b).Encode(v)
	//if err != nil {
	//	log.Fatal("NewEncoder failed:", err.Error())
	//}
	//return b.Bytes()
}

func OsUnmarshal(data []byte, v interface{}) {
	err := json.Unmarshal(data, v)
	//b := bytes.NewBuffer(data)
	//err := gob.NewDecoder(b).Decode(v)
	if err != nil {
		fmt.Println(string(data))
		//str := err.Error()
		//fmt.Println(str[err.offset:1])
		log.Fatal("OsUnmarshal failed:", err.Error())
	}
}

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

func OsRoundHalf(v float64) float64 {
	return OsRoundDown(v + OsTrnFloat(v < 0, -0.5, 0.5))
}

type OsV2f struct {
	X float32
	Y float32
}

func (a OsV2f) Add(b OsV2f) OsV2f {
	return OsV2f{a.X + b.X, a.Y + b.Y}
}
func (a OsV2f) Sub(b OsV2f) OsV2f {
	return OsV2f{a.X - b.X, a.Y - b.Y}
}
func (a OsV2f) Mul(b OsV2f) OsV2f {
	return OsV2f{a.X * b.X, a.Y * b.Y}
}
func (a OsV2f) Div(b OsV2f) OsV2f {
	return OsV2f{a.X / b.X, a.Y / b.Y}
}
func (a OsV2f) MulV(t float32) OsV2f {
	return OsV2f{a.X * t, a.Y * t}
}
func (a OsV2f) DivV(t float32) OsV2f {
	return a.MulV(1 / t)
}
func (a OsV2f) toV2() OsV2 {
	return OsV2{int(a.X), int(a.Y)}
}
func (a OsV2f) Cmp(b OsV2f) bool {
	return a.X == b.X && a.Y == b.Y
}
func (a OsV2f) Min(b OsV2f) OsV2f {
	return OsV2f{OsMinFloat32(a.X, b.X), OsMinFloat32(a.Y, b.Y)}
}
func (a OsV2f) Max(b OsV2f) OsV2f {
	return OsV2f{OsMaxFloat32(a.X, b.X), OsMaxFloat32(a.Y, b.Y)}
}
func (v OsV2f) Len() float32 {
	return float32(math.Sqrt(float64(v.X*v.X + v.Y*v.Y)))
}

type OsV2 struct {
	X int
	Y int
}

func OsV2_32(x, y int32) OsV2 {
	return OsV2{int(x), int(y)}
}

func (v *OsV2) Get32() (int32, int32) {
	return int32(v.X), int32(v.Y)
}

func (v *OsV2) EqAdd(val OsV2) {
	v.X += val.X
	v.Y += val.Y
}
func (v *OsV2) EqSub(vel OsV2) {
	v.X -= vel.X
	v.Y -= vel.Y
}

func (a OsV2) Add(b OsV2) OsV2 {
	return OsV2{a.X + b.X, a.Y + b.Y}
}
func (a OsV2) Sub(b OsV2) OsV2 {
	return OsV2{a.X - b.X, a.Y - b.Y}
}
func (a OsV2) MulV(t float32) OsV2 {
	return OsV2{int(float32(a.X) * t), int(float32(a.Y) * t)}
}

func (a OsV2) Aprox(b OsV2, t float32) OsV2 {
	return a.Add(b.Sub(a).MulV(t))
}

func (a OsV2) toV2f() OsV2f {
	return OsV2f{float32(a.X), float32(a.Y)}
}

func (a OsV2) Cmp(b OsV2) bool {
	return a.X == b.X && a.Y == b.Y
}

func (start OsV2) Inside(end OsV2, test OsV2) bool {
	return test.X >= start.X && test.Y >= start.Y && test.X < end.X && test.Y < end.Y
}

func (a OsV2) Min(b OsV2) OsV2 {
	return OsV2{OsMin(a.X, b.X), OsMin(a.Y, b.Y)}
}

func (a OsV2) Max(b OsV2) OsV2 {
	return OsV2{OsMax(a.X, b.X), OsMax(a.Y, b.Y)}
}

func (v OsV2) Is() bool {
	return v.X != 0 && v.Y != 0
}

func (v OsV2) IsZero() bool {
	return v.X == 0 && v.Y == 0
}

func (v *OsV2) Switch() {
	*v = OsV2{v.Y, v.X}
}

func (v *OsV2) Sort() {
	if v.X > v.Y {
		v.Switch()
	}
}

func (v OsV2) Len() float32 {
	return float32(math.Sqrt(float64(v.X*v.X + v.Y*v.Y)))
}

func (v OsV2) Angle() float32 {
	return float32(math.Atan2(float64(v.Y), float64(v.X))) //<-PI, PI>
}

func (a OsV2) Distance(b OsV2) float32 {
	return a.Sub(b).Len()
}

func OsV2_Intersect(a OsV2, b OsV2) OsV2 {
	v := OsV2{OsMax(a.X, b.X), OsMax(a.Y, b.Y)}

	if v.X > v.Y {
		return OsV2{}
	}
	return v
}

func OsV2_InRatio(rect OsV2, orig OsV2) OsV2 {
	rectRatio := float32(rect.X) / float32(rect.Y)
	origRatio := float32(orig.X) / float32(orig.Y)

	var ratio float32
	if origRatio > rectRatio {
		ratio = float32(rect.X) / float32(orig.X)
	} else {
		ratio = float32(rect.Y) / float32(orig.Y)
	}
	return orig.MulV(ratio)
}

func OsV2_OutRatio(rect OsV2, orig OsV2) OsV2 {
	rectRatio := float32(rect.X) / float32(rect.Y)
	origRatio := float32(orig.X) / float32(orig.Y)

	var ratio float32
	if origRatio < rectRatio {
		ratio = float32(rect.X) / float32(orig.X)
	} else {
		ratio = float32(rect.Y) / float32(orig.Y)
	}
	return orig.MulV(ratio)
}

func (coord OsV4) Align(size OsV2, align OsV2) OsV2 {
	start := coord.Start

	if align.X == 0 {
		// left
	} else if align.X == 1 {
		// center
		if size.X > coord.Size.X {
			start.X = coord.Start.X // + H / 2
		} else {
			start.X = coord.Middle().X - size.X/2
		}
	} else {
		// right
		start.X = coord.End().X - size.X
	}

	// y
	if size.Y >= coord.Size.Y {
		start.Y += (coord.Size.Y - size.Y) / 2
	} else {
		if align.Y == 0 {
			start.Y = coord.Start.Y // + H / 2
		} else if align.Y == 1 {
			start.Y += (coord.Size.Y - size.Y) / 2
		} else if align.Y == 2 {
			start.Y += (coord.Size.Y) - size.Y
		}
	}

	return start
}

type OsV4 struct {
	Start OsV2
	Size  OsV2
}

func InitOsV4(x, y, w, h int) OsV4 {
	return OsV4{OsV2{x, y}, OsV2{w, h}}
}

func InitOsV4Mid(mid OsV2, size OsV2) OsV4 {
	return InitOsV4(mid.X-size.X/2, mid.Y-size.Y/2, size.X, size.Y)
}

func InitOsV4ab(a OsV2, b OsV2) OsV4 {
	st := OsV2{OsMin(a.X, b.X), OsMin(a.Y, b.Y)}
	sz := OsV2{OsAbs(a.X - b.X), OsAbs(a.Y - b.Y)}
	return InitOsV4(st.X, st.Y, sz.X, sz.Y)
}

func (v OsV4) End() OsV2 {
	return OsV2{v.Start.X + v.Size.X, v.Start.Y + v.Size.Y}
}

func (v OsV4) Is() bool {
	return v.Size.Is()
}

func (v OsV4) IsZero() bool {
	return v.Size.IsZero()
}

func (v OsV4) Check() {
	if v.Size.X < 0 {
		v.Size.X = 0
	}
	if v.Size.Y < 0 {
		v.Size.Y = 0
	}
}

func (v OsV4) GetPos(x, y float64) OsV2 {
	return OsV2{v.Start.X + int(float64(v.Size.X)*x), v.Start.Y + int(float64(v.Size.Y)*y)}
}

func (q OsV4) CropX(space int) OsV4 {
	space *= 2
	if space > q.Size.X {
		space = q.Size.X
	}
	return InitOsV4(q.Start.X+space/2, q.Start.Y, q.Size.X-space, q.Size.Y)
}

func (q OsV4) CropY(space int) OsV4 {
	space *= 2
	if space > q.Size.Y {
		space = q.Size.Y
	}
	return InitOsV4(q.Start.X, q.Start.Y+space/2, q.Size.X, q.Size.Y-space)
}

func (q OsV4) Crop(space int) OsV4 {
	r := q.CropX(space)
	return r.CropY(space)
}

func (q OsV4) Inner(top, bottom, left, right int) OsV4 {
	for q.Size.X < (left + right) { //for!
		left--
		right--
	}
	for q.Size.Y < (top + bottom) { //for!
		top--
		bottom--
	}
	return InitOsV4(q.Start.X+left, q.Start.Y+top, q.Size.X-(left+right), q.Size.Y-(top+bottom))
}

func (v OsV4) Middle() OsV2 {
	return v.Start.Add(v.Size.MulV(0.5))
}

func (v OsV4) Inside(test OsV2) bool {
	return v.Start.Inside(v.End(), test)
}
func (a OsV4) Cmp(b OsV4) bool {
	return a.Start.Cmp(b.Start) && a.Size.Cmp(b.Size)
}

func OsV4_center(out OsV4, in OsV2) OsV4 {
	r := OsV4{out.Start, in}

	if out.Size.X > in.X {
		r.Start.X += (out.Size.X - in.X) / 2
	}
	if out.Size.Y > in.Y {
		r.Start.Y += (out.Size.Y - in.Y) / 2
	}
	return r
}

func OsV4_centerFull(out OsV4, in OsV2) OsV4 {
	r := OsV4{out.Start, in}

	if out.Size.X != in.X {
		r.Start.X += (out.Size.X - in.X) / 2
	}
	if out.Size.Y != in.Y {
		r.Start.Y += (out.Size.Y - in.Y) / 2
	}
	return r
}

func (a OsV4) Area() int {
	return a.Size.X * a.Size.Y
}

func (a OsV4) Extend(b OsV4) OsV4 {

	start := OsV2{OsMin(a.Start.X, b.Start.X), OsMin(a.Start.Y, b.Start.Y)}

	ae := a.End()
	be := b.End()

	end := OsV2{OsMax(ae.X, be.X), OsMax(ae.Y, be.Y)}

	return OsV4{start, end.Sub(start)}
}

func (a OsV4) Extend2(q OsV4, v OsV2) OsV4 {

	start := OsV2{OsMin(q.Start.X, v.X), OsMin(q.Start.Y, v.Y)}

	end := q.End()
	end.X = OsMax(end.X, v.X)
	end.Y = OsMax(end.Y, v.Y)

	return OsV4{start, end.Sub(start)}
}

func (a OsV4) HasCover(b OsV4) bool {
	q := a.Extend(b)
	return q.Size.X < (a.Size.X+b.Size.X) && q.Size.Y < (a.Size.Y+b.Size.Y)
}

func (qA OsV4) GetIntersect(qB OsV4) OsV4 {

	if qA.HasCover(qB) {
		v_start := qA.Start.Max(qB.Start)
		v_end := qA.End().Min(qB.End())

		return OsV4{v_start, v_end.Sub(v_start)}
	}
	return OsV4{Start: qA.Start}
}

func (qA OsV4) HasIntersect(qB OsV4) bool {

	q := qA.GetIntersect(qB)
	return q.Is()
}

func OsV4_relativeSurround(src OsV4, dst OsV4, screen OsV4, priorUp bool) OsV4 {

	q := dst
	q.Start = q.Start.Sub(screen.Start)

	srcStart := src.Start.Sub(screen.Start)
	srcSize := src.Size

	up := srcStart.Y > (screen.Size.Y - srcStart.Y - srcSize.Y)
	if !up && priorUp {
		up = srcStart.Y > q.Size.Y //check enough space
	}

	q.Start.X = srcStart.X
	if q.Start.X+q.Size.X > screen.Size.X {
		q.Start.X = screen.Size.X - q.Size.X //move to left
	}

	if up {
		q.Start.Y = srcStart.Y - q.Size.Y
		if q.Start.Y < 0 {
			q.Size.Y = srcStart.Y
			q.Start.Y = 0
		}
	} else {
		q.Start.Y = srcStart.Y + srcSize.Y
		if q.Start.Y+q.Size.Y > screen.Size.Y {
			q.Size.Y = screen.Size.Y - q.Start.Y
		}
	}

	q.Start = q.Start.Add(screen.Start)
	return q
}

func (v *OsV4) RelativePos(p OsV2) OsV2f {
	s := p.Sub(v.Start)
	return OsV2f{float32(s.X) / float32(v.Size.X), float32(s.Y) / float32(v.Size.Y)}
}

func (v *OsV4) Relative(q OsV4) (x, y, w, h float32) {
	s := v.RelativePos(q.Start)
	e := v.RelativePos(q.End())
	return s.X, s.Y, (e.X - s.X), (e.Y - s.Y)
}

func (v OsV4) Cut(x, y, w, h float64) OsV4 {

	return InitOsV4(
		v.Start.X+int(float64(v.Size.X)*x),
		v.Start.Y+int(float64(v.Size.Y)*y),
		int(float64(v.Size.X)*w),
		int(float64(v.Size.Y)*h))
}

func (v OsV4) CutEx(x, y, w, h float64, space, spaceX, spaceY int) OsV4 {
	v = v.CropX(spaceX)
	v = v.CropY(spaceY)
	v = v.Crop(space)
	v = v.Cut(x, y, w, h)
	return v
}

func OsFileGetNameWithoutExt(fileName string) string {

	fileName = filepath.Base(fileName)
	fileName = strings.TrimSuffix(fileName, filepath.Ext(fileName))
	return fileName
}

func OsFileExists(fileName string) bool {
	info, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func OsFolderUniqueNewName(folder, name string) (string, string) {
	folder, _ = strings.CutSuffix(folder, "/")

	path := folder + "/" + name
	if !OsFolderExists(path) {
		return name, path
	}

	i := 2
	for {
		new_name := name + "_" + strconv.Itoa(i)
		new_path := folder + "/" + new_name
		if !OsFolderExists(new_path) {
			return new_name, new_path
		}
		i++
	}
}

func OsFolderExists(fileName string) bool {
	info, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func OsFileTime(fileName string) int64 {
	info, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return -1
	}
	return info.ModTime().UnixNano()
}

func OsFolderCreate(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}

func OsFolderRemove(path string) error {
	return os.RemoveAll(path)
}
func OsFileRemove(path string) error {
	return os.Remove(path)
}
func OsFileRename(path string, newPath string) error {
	return os.Rename(path, newPath)
}

func OsFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return -1, err
	}
	return info.Size(), nil
}

func OsFileCopy(src, dst string) error {
	srcFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !srcFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	if err != nil {
		return err
	}

	err = os.Chmod(dst, srcFileStat.Mode())
	if err != nil {
		return err
	}

	return err
}

func OsFolderCopy(src, dst string) error {
	src, _ = strings.CutSuffix(src, "/")
	dst, _ = strings.CutSuffix(dst, "/")

	err := OsFolderCreate(dst)
	if err != nil {
		return err
	}

	dir, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, it := range dir {
		if it.IsDir() {
			err = OsFolderCopy(src+"/"+it.Name(), dst+"/"+it.Name())
			if err != nil {
				return err
			}
		} else {
			OsFileCopy(src+"/"+it.Name(), dst+"/"+it.Name())
		}
	}

	return nil
}

func OsFolderHash(folder string, ignoreNames []string) (int64, error) {

	dir, err := os.ReadDir(folder)
	if err != nil {
		return 0, err
	}

	h := sha256.New()

	for _, file := range dir {
		found := false
		for _, nm := range ignoreNames {
			if nm == file.Name() {
				found = true
				break
			}
		}
		if found {
			continue //skip
		}

		var tm [8]byte
		binary.LittleEndian.PutUint64(tm[:], uint64(OsFileTime(filepath.Join(folder, file.Name()))))

		h.Write([]byte(file.Name()))
		h.Write(tm[:])
	}

	return int64(binary.LittleEndian.Uint64(h.Sum(nil))), nil
}

const g_special = "ěščřžťďňĚŠČŘŽŤĎŇáíýéóúůÁÍÝÉÓÚŮÄäÖöÜüẞß"

func OsIsTextWord(ch rune) bool {
	return ((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || (ch == '_')) || strings.ContainsRune(g_special, ch)
}

func OsUlit_GetUID() (string, error) {

	device, err := os.Hostname()
	if err != nil {
		return "", err
	}

	h, err := InitOsHash([]byte(device))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.h), nil
}

const OsHash_SIZE = 32

type OsHash struct {
	h [OsHash_SIZE]byte
}

func InitOsHash(src []byte) (OsHash, error) {

	if len(src) == 0 {
		return OsHash{}, nil //zeros
	}

	h := sha256.New()
	_, err := h.Write(src)
	if err != nil {
		return OsHash{}, err
	}

	return OsHash{h: [OsHash_SIZE]byte(h.Sum(nil))}, nil
}

func InitOsHashCopy(src_hash []byte) OsHash {
	var h OsHash
	if len(src_hash) == OsHash_SIZE {
		copy(h.h[:], src_hash[:])
	}
	//else empty
	return h
}

func (a *OsHash) Cmp(b *OsHash) bool {
	return bytes.Equal(a.h[:], b.h[:])
}

func (a OsHash) CmpBytes(b []byte) bool {
	if len(b) == OsHash_SIZE {
		return bytes.Equal(a.h[:], b)
	}

	return a.Cmp(&OsHash{}) //is empty
}

func (h *OsHash) GetInt64() int64 {
	return int64(binary.LittleEndian.Uint64(h.h[:]))
}

func (h *OsHash) Hex() string {
	return hex.EncodeToString(h.h[:])
}

type OsFileList struct {
	Name  string
	IsDir bool
	Subs  []OsFileList
}

func OsFileListBuild(path string, name string, isDir bool) OsFileList {
	var fl OsFileList
	fl.Name = name
	fl.IsDir = isDir

	if isDir {
		dir, err := os.ReadDir(path)
		if err == nil {
			for _, file := range dir {
				fl.Subs = append(fl.Subs, OsFileListBuild(path+"/"+file.Name(), file.Name(), file.IsDir()))
			}
		}
	}
	return fl
}

func (fl *OsFileList) FindInSubs(name string, isDir bool) int {
	for i, f := range fl.Subs {
		if f.Name == name && f.IsDir == isDir {
			return i
		}
	}
	return -1
}

func OsText_JSONtoRAW(str string) (string, error) {
	if str != "" && str[0] != '"' {
		fmt.Printf("Error: String(%s) doesn't start with quotes\n", str)
	}

	//str = strings.Clone(str)
	v, err := strconv.Unquote(str)
	if err != nil {
		return "", err
	}
	return v, nil
}

func OsText_RAWtoJSON(str string) string {
	//str = strings.Clone(str)
	v := strconv.Quote(str)
	return v
}

func OsText_InterfaceToJSON(value interface{}) string {
	switch vv := value.(type) {
	case string:
		return OsText_RAWtoJSON(vv)
	case float64:
		return strconv.FormatFloat(vv, 'f', -1, 64)

	case int: //this is just for SANodeAttr init, otherwise float64
		return strconv.Itoa(vv)

	case []byte: //this is just for SANodeAttr init, otherwise OsBlob
		if len(vv) > 0 && (vv[0] == '{' || vv[0] == '[') {
			return string(vv) //map or array
		}
		return "\"" + string(vv) + "\"" //binary/hex
	case OsBlob:
		if len(vv.data) > 0 && (vv.data[0] == '{' || vv.data[0] == '[') {
			return string(vv.data) //map or array
		}
		return "\"" + string(vv.data) + "\"" //binary/hex

	default:
		fmt.Println("Warning: Unknown SAValue conversion into String")
	}
	return "\"\""
}

func Os_StartProfile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	pprof.StartCPUProfile(f)
	return nil
}
func Os_StopProfile() {
	pprof.StopCPUProfile()
}

func Os_GetAmountRAM() (uint64, uint64) {
	return memory.TotalMemory(), memory.FreeMemory()
}

type OsBlob struct {
	data []byte
	hash OsHash
}

func InitOsBlob(blob []byte) OsBlob {
	var b OsBlob
	b.data = blob
	b.hash, _ = InitOsHash(blob) //err ...
	return b
}

func (b *OsBlob) Hex() string {
	return hex.EncodeToString(b.data)
}

func (b *OsBlob) Len() int {
	return len(b.data)
}

func (a *OsBlob) CmpHash(b *OsBlob) bool {
	return a.hash.Cmp(&b.hash)
}

func OsNextPowOf2(n int) int {
	k := 1
	for k < n {
		k = k << 1
	}
	return k
}

type OsWriterSeeker struct {
	buf bytes.Buffer
	pos int
}

func (wr *OsWriterSeeker) Write(p []byte) (n int, err error) {
	if extra := wr.pos - wr.buf.Len(); extra > 0 {
		if _, err := wr.buf.Write(make([]byte, extra)); err != nil {
			return n, err
		}
	}

	if wr.pos < wr.buf.Len() {
		n = copy(wr.buf.Bytes()[wr.pos:], p)
		p = p[n:]
	}

	if len(p) > 0 {
		var bt int
		bt, err = wr.buf.Write(p)
		n += bt
	}

	wr.pos += n
	return n, err
}
func (wr *OsWriterSeeker) Seek(offset int64, whence int) (int64, error) {
	newPos, offs := 0, int(offset)
	switch whence {
	case io.SeekStart:
		newPos = offs
	case io.SeekCurrent:
		newPos = wr.pos + offs
	case io.SeekEnd:
		newPos = wr.buf.Len() + offs
	}
	if newPos < 0 {
		return 0, errors.New("pos is negative")
	}
	wr.pos = newPos
	return int64(newPos), nil
}
func (wr *OsWriterSeeker) Reader() io.Reader {
	return bytes.NewReader(wr.buf.Bytes())
}
func (wr *OsWriterSeeker) Close() error {
	return nil
}
func (wr *OsWriterSeeker) BytesReader() *bytes.Reader {
	return bytes.NewReader(wr.buf.Bytes())
}

func OsConvertBytesToString(bytes int) string {
	var str string
	if bytes >= 1000*1000*1000 {
		str = fmt.Sprintf("%.1f GB", float64(bytes)/(1000*1000*1000))
	} else if bytes >= 1000*1000 {
		str = fmt.Sprintf("%.1f MB", float64(bytes)/(1000*1000))
	} else if bytes >= 1000 {
		str = fmt.Sprintf("%.1f KB", float64(bytes)/(1000))
	} else {
		str = fmt.Sprintf("%d B", bytes)
	}
	return str
}

func OsGetStringStartsWithUpper(str string) string {
	if len(str) == 0 {
		return ""
	}
	return strings.ToUpper(str[0:1]) + str[1:]
}

type SA_Drop_POS int

const (
	SA_Drop_INSIDE  SA_Drop_POS = 0
	SA_Drop_V_LEFT  SA_Drop_POS = 1
	SA_Drop_V_RIGHT SA_Drop_POS = 2
	SA_Drop_H_LEFT  SA_Drop_POS = 3
	SA_Drop_H_RIGHT SA_Drop_POS = 4
)

func OsMoveElementIndex(src int, dst int, pos SA_Drop_POS) int {
	//check
	if src < dst && (pos == SA_Drop_V_LEFT || pos == SA_Drop_H_LEFT) {
		dst--
	}
	if src > dst && (pos == SA_Drop_V_RIGHT || pos == SA_Drop_H_RIGHT) {
		dst++
	}
	return dst
}

func OsMoveElement[T any](array_src *[]T, array_dst *[]T, src int, dst int) {
	//move(by swap one-by-one)
	if array_src == array_dst {
		for i := src; i < dst; i++ {
			(*array_dst)[i], (*array_dst)[i+1] = (*array_dst)[i+1], (*array_dst)[i]
		}
		for i := src; i > dst; i-- {
			(*array_dst)[i], (*array_dst)[i-1] = (*array_dst)[i-1], (*array_dst)[i]
		}
	} else {
		backup := (*array_src)[src]

		//remove
		*array_src = append((*array_src)[:src], (*array_src)[src+1:]...)

		//insert
		if dst < len(*array_dst) {
			*array_dst = append((*array_dst)[:dst+1], (*array_dst)[dst:]...)
			(*array_dst)[dst] = backup
		} else {
			*array_dst = append(*array_dst, backup)
			dst = len(*array_dst) - 1
		}
	}
}
