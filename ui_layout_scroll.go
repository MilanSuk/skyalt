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

import "image/color"

type UiLayoutScroll struct {
	wheel int // pixel move

	data_height   int
	screen_height int

	clickRel int

	timeWheel int64

	Show   bool
	Narrow bool

	attach *UiLayoutScroll
}

func (scroll *UiLayoutScroll) Init() {

	scroll.clickRel = 0
	scroll.wheel = 0
	scroll.data_height = 1
	scroll.screen_height = 1
	scroll.timeWheel = 0
	scroll.Show = true
	scroll.Narrow = false
}

func (scroll *UiLayoutScroll) _getWheel(wheel int) int {

	if scroll.data_height > scroll.screen_height {
		return OsClamp(wheel, 0, (scroll.data_height - scroll.screen_height))
	}
	return 0
}

func (scroll *UiLayoutScroll) IsDown() bool {
	return (scroll.wheel + scroll.screen_height) == scroll.data_height
}

func (scroll *UiLayoutScroll) GetWheel() int {
	return scroll._getWheel(scroll.wheel)
}

func (scroll *UiLayoutScroll) SetWheel(wheelPixel int) bool {
	oldWheel := scroll.wheel

	scroll.wheel = wheelPixel
	scroll.wheel = scroll.GetWheel() //clamp by boundaries

	if oldWheel != scroll.wheel {
		scroll.timeWheel = OsTicks()

		if scroll.attach != nil {
			scroll.attach.wheel = scroll.wheel
		}
	}

	return oldWheel != scroll.wheel
}

func (scroll *UiLayoutScroll) Is() bool {
	return scroll.Show && scroll.data_height > scroll.screen_height
}

func (scroll *UiLayoutScroll) GetScrollBackCoordV(view OsV4, ui *Ui) OsV4 {
	WIDTH := scroll._GetWidth(ui)
	H := 0 // OsTrn(shorter, WIDTH, 0)
	return OsV4{OsV2{view.Start.X + view.Size.X, view.Start.Y}, OsV2{WIDTH, scroll.screen_height - H}}
}
func (scroll *UiLayoutScroll) GetScrollBackCoordH(view OsV4, ui *Ui) OsV4 {
	WIDTH := scroll._GetWidth(ui)
	H := 0 //OsTrn(shorter, WIDTH, 0)
	return OsV4{OsV2{view.Start.X, view.Start.Y + view.Size.Y}, OsV2{scroll.screen_height - H, WIDTH}}
}

func (scroll *UiLayoutScroll) _GetWidth(ui *Ui) int {
	widthWin := ui.Cell() / 2
	if scroll.Narrow {
		return OsMax(4, widthWin/10)
	}
	return widthWin
}

func (scroll *UiLayoutScroll) _UpdateV(view OsV4, ui *Ui) OsV4 {
	if !scroll.Show {
		scroll.wheel = 0
	}

	var outSlider OsV4
	if scroll.data_height <= scroll.screen_height {
		outSlider.Start = view.Start

		outSlider.Size.X = scroll._GetWidth(ui)
		outSlider.Size.Y = view.Size.Y // self.screen_height
	} else {
		outSlider.Start.X = view.Start.X
		outSlider.Start.Y = view.Start.Y + int(float64(view.Size.Y)*(float64(scroll.GetWheel())/float64(scroll.data_height)))

		outSlider.Size.X = scroll._GetWidth(ui)
		outSlider.Size.Y = int(float64(view.Size.Y) * (float64(scroll.screen_height) / float64(scroll.data_height)))
	}
	return outSlider
}

func (scroll *UiLayoutScroll) _UpdateH(start OsV2, ui *Ui) OsV4 {
	if !scroll.Show {
		scroll.wheel = 0
	}

	var outSlider OsV4
	if scroll.data_height <= scroll.screen_height {
		outSlider.Start = start

		outSlider.Size.X = scroll.screen_height
		outSlider.Size.Y = scroll._GetWidth(ui)
	} else {
		outSlider.Start.X = start.X + int(float64(scroll.screen_height)*(float64(scroll.GetWheel())/float64(scroll.data_height)))
		outSlider.Start.Y = start.Y

		outSlider.Size.X = int(float64(scroll.screen_height) * (float64(scroll.screen_height) / float64(scroll.data_height)))
		outSlider.Size.Y = scroll._GetWidth(ui)
	}
	return outSlider
}

func (scroll *UiLayoutScroll) _GetSlideCd(ui *Ui) color.RGBA {

	cd_slide := ui.sync.GetPalette().GetGrey(0.5)
	if scroll.data_height <= scroll.screen_height {
		cd_slide = Color_Aprox(ui.sync.GetPalette().OnB, cd_slide, 0.5) // disable
	}

	return cd_slide
}

func (scroll *UiLayoutScroll) DrawV(view OsV4, showBackground bool, ui *Ui) {
	slider := scroll._UpdateV(view, ui)

	slider = slider.Crop(OsMax(1, slider.Size.X/5))

	// make scroll visible if there is a lot of records(items)
	if slider.Size.Y == 0 {
		c := ui.Cell() / 4
		slider.Start.Y -= c / 2
		slider.Size.Y += c
	}

	if showBackground {
		cd := ui.sync.GetPalette().OnB
		cd.A = 30
		ui.GetWin().buff.AddRect(view, cd, 0)
	}
	cd := scroll._GetSlideCd(ui)
	rounding := OsRoundUp(float64(slider.Size.X) * ui.sync.GetRounding())
	ui.GetWin().buff.AddRectRound(slider, rounding, cd, 0)
}

func (scroll *UiLayoutScroll) DrawH(view OsV4, showBackground bool, ui *Ui) {
	slider := scroll._UpdateH(view.Start, ui)

	slider = slider.Crop(OsMax(1, slider.Size.Y/5))

	// make scroll visible if there is a lot of records(items)
	if slider.Size.Y == 0 {

		c := ui.Cell() / 4
		slider.Start.X -= c / 2
		slider.Size.X += c
	}

	if showBackground {
		cd := ui.sync.GetPalette().OnB
		cd.A = 30
		ui.GetWin().buff.AddRect(view, cd, 0)
	}

	cd := scroll._GetSlideCd(ui)
	rounding := OsRoundUp(float64(slider.Size.Y) * ui.sync.GetRounding())
	ui.GetWin().buff.AddRectRound(slider, rounding, cd, 0)
}

func (scroll *UiLayoutScroll) _GetTempScroll(srcl int, ui *Ui) int {
	return ui.Cell() * srcl
}

func (scroll *UiLayoutScroll) IsMove(layout *Layout, wheel_add int, deep int, ver bool) bool {

	inside := layout.CropWithScroll().Inside(layout.ui.GetWin().io.Touch.Pos)
	if inside {
		//test childs
		for _, div := range layout.childs {
			if div.IsShown() {
				if ver && div.scrollV.IsMove(div, wheel_add, deep+1, ver) {
					return deep > 0 //bottom layer must return false(can't scroll, because upper layer can scroll)
				}
				if !ver && div.scrollH.IsMove(div, wheel_add, deep+1, ver) {
					return deep > 0 //bottom layer must return false(can't scroll, because upper layer can scroll)
				}
			}
		}

		if scroll.Is() {
			curr := scroll.GetWheel()
			return scroll._getWheel(curr+wheel_add) != curr //can move => true
		}
	}

	return false
}

func (scroll *UiLayoutScroll) TouchV(layout *Layout) bool {
	backup_wheel := scroll.wheel
	ui := layout.ui

	canUp := scroll.IsMove(layout, -1, 0, true)
	canDown := scroll.IsMove(layout, +1, 0, true)
	if ui.GetWin().io.Touch.Wheel != 0 && !ui.GetWin().io.Keys.Shift {
		if (ui.GetWin().io.Touch.Wheel < 0 && canUp) || (ui.GetWin().io.Touch.Wheel > 0 && canDown) {
			if scroll.SetWheel(scroll.GetWheel() + scroll._GetTempScroll(ui.GetWin().io.Touch.Wheel, ui)) {
				ui.GetWin().io.Touch.Wheel = 0 // let child scroll
			}
		}
	}

	if !layout.ui.edit.IsActive() && !layout.ui.touch.IsActive() && !ui.GetWin().io.Keys.Shift {
		if ui.GetWin().io.Keys.ArrowU && canUp {
			if scroll.SetWheel(scroll.GetWheel() - ui.Cell()) {
				ui.GetWin().io.Keys.ArrowU = false
			}
		}
		if ui.GetWin().io.Keys.ArrowD && canDown {
			if scroll.SetWheel(scroll.GetWheel() + ui.Cell()) {
				ui.GetWin().io.Keys.ArrowD = false
			}
		}

		if ui.GetWin().io.Keys.Home && canUp {
			if scroll.SetWheel(0) {
				ui.GetWin().io.Keys.Home = false
			}
		}
		if ui.GetWin().io.Keys.End && canDown {
			if scroll.SetWheel(scroll.data_height) {
				ui.GetWin().io.Keys.End = false
			}
		}

		if ui.GetWin().io.Keys.PageU && canUp {
			if scroll.SetWheel(scroll.GetWheel() - scroll.screen_height) {
				ui.GetWin().io.Keys.PageU = false
			}
		}
		if ui.GetWin().io.Keys.PageD && canDown {
			if scroll.SetWheel(scroll.GetWheel() + scroll.screen_height) {
				ui.GetWin().io.Keys.PageD = false
			}
		}
	}

	if !scroll.Is() {
		return backup_wheel != scroll.wheel
	}

	scrollCoord := layout.scrollV.GetScrollBackCoordV(layout.view, ui)

	sliderFront := scroll._UpdateV(scrollCoord, ui)
	midSlider := sliderFront.Size.Y / 2

	isTouched := layout.ui.touch.IsFnMove(0, layout.UID, 0, 0)
	if ui.GetWin().io.Touch.Start {
		isTouched = sliderFront.Inside(ui.GetWin().io.Touch.Pos)
		scroll.clickRel = ui.GetWin().io.Touch.Pos.Y - sliderFront.Start.Y - midSlider // rel to middle of front slide
	}

	if isTouched { // click on slider
		mid := float64((ui.GetWin().io.Touch.Pos.Y - scrollCoord.Start.Y) - midSlider - scroll.clickRel)
		scroll.SetWheel(int((mid / float64(scrollCoord.Size.Y)) * float64(scroll.data_height)))

	} else if ui.GetWin().io.Touch.Start && scrollCoord.Inside(ui.GetWin().io.Touch.Pos) && !sliderFront.Inside(ui.GetWin().io.Touch.Pos) { // click(once) on background
		mid := float64((ui.GetWin().io.Touch.Pos.Y - scrollCoord.Start.Y) - midSlider)
		scroll.SetWheel(int((mid / float64(scrollCoord.Size.Y)) * float64(scroll.data_height)))

		// switch to 'click on slider'
		isTouched = true
		scroll.clickRel = 0
	}

	if isTouched {
		layout.ui.touch.Set(0, layout.UID, 0, 0)
	}

	scroll.attach = nil //reset

	return backup_wheel != scroll.wheel
}

func (scroll *UiLayoutScroll) TouchH(needShiftWheel bool, layout *Layout) bool {
	backup_wheel := scroll.wheel
	ui := layout.ui

	canLeft := scroll.IsMove(layout, -1, 0, !ui.GetWin().io.Keys.Shift)
	canRight := scroll.IsMove(layout, +1, 0, !ui.GetWin().io.Keys.Shift)
	if ui.GetWin().io.Touch.Wheel != 0 && (!needShiftWheel || ui.GetWin().io.Keys.Shift) {
		if (ui.GetWin().io.Touch.Wheel < 0 && canLeft) || (ui.GetWin().io.Touch.Wheel > 0 && canRight) {
			if scroll.SetWheel(scroll.GetWheel() + scroll._GetTempScroll(ui.GetWin().io.Touch.Wheel, ui)) {
				ui.GetWin().io.Touch.Wheel = 0 // let child scroll
			}
		}
	}

	if !layout.ui.edit.IsActive() && !layout.ui.touch.IsActive() && (!needShiftWheel || ui.GetWin().io.Keys.Shift) {
		if ui.GetWin().io.Keys.ArrowL && canLeft {
			if scroll.SetWheel(scroll.GetWheel() - ui.Cell()) {
				ui.GetWin().io.Keys.ArrowL = false
			}
		}
		if ui.GetWin().io.Keys.ArrowR && canRight {
			if scroll.SetWheel(scroll.GetWheel() + ui.Cell()) {
				ui.GetWin().io.Keys.ArrowR = false
			}
		}

		if ui.GetWin().io.Keys.Home && canLeft {
			if scroll.SetWheel(0) {
				ui.GetWin().io.Keys.Home = false
			}
		}
		if ui.GetWin().io.Keys.End && canRight {
			if scroll.SetWheel(scroll.data_height) {
				ui.GetWin().io.Keys.End = false
			}
		}

		if ui.GetWin().io.Keys.PageU && canLeft {
			if scroll.SetWheel(scroll.GetWheel() - scroll.screen_height) {
				ui.GetWin().io.Keys.PageU = false
			}
		}
		if ui.GetWin().io.Keys.PageD && canRight {
			if scroll.SetWheel(scroll.GetWheel() + scroll.screen_height) {
				ui.GetWin().io.Keys.PageD = false
			}
		}
	}

	if !scroll.Is() {
		return backup_wheel != scroll.wheel
	}

	scrollCoord := layout.scrollV.GetScrollBackCoordH(layout.view, ui)

	sliderFront := scroll._UpdateH(scrollCoord.Start, ui)
	midSlider := sliderFront.Size.X / 2

	isTouched := layout.ui.touch.IsFnMove(0, 0, layout.UID, 0)
	if ui.GetWin().io.Touch.Start {
		isTouched = sliderFront.Inside(ui.GetWin().io.Touch.Pos)
		scroll.clickRel = ui.GetWin().io.Touch.Pos.X - sliderFront.Start.X - midSlider // rel to middle of front slide
	}

	if isTouched { // click on slider
		mid := float64((ui.GetWin().io.Touch.Pos.X - scrollCoord.Start.X) - midSlider - scroll.clickRel)
		scroll.SetWheel(int((mid / float64(scroll.screen_height)) * float64(scroll.data_height)))

	} else if ui.GetWin().io.Touch.Start && scrollCoord.Inside(ui.GetWin().io.Touch.Pos) && !sliderFront.Inside(ui.GetWin().io.Touch.Pos) { // click(once) on background
		mid := float64((ui.GetWin().io.Touch.Pos.X - scrollCoord.Start.X) - midSlider)
		scroll.SetWheel(int((mid / float64(scroll.screen_height)) * float64(scroll.data_height)))

		// switch to 'click on slider'
		isTouched = true
		scroll.clickRel = 0
	}

	if isTouched {
		layout.ui.touch.Set(0, 0, layout.UID, 0)
	}

	return backup_wheel != scroll.wheel
}

func (scroll *UiLayoutScroll) TryDragScroll(fast_dt int, sign int, ui *Ui) bool {
	wheelOld := scroll.GetWheel()

	dt := int64((1.0 / 2.0) / float64(fast_dt) * 1000)

	if OsTicks()-scroll.timeWheel > dt {
		scroll.SetWheel(scroll.GetWheel() + scroll._GetTempScroll(sign, ui))
	}

	return scroll.GetWheel() != wheelOld
}
