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
	"log"
	"time"
)

func main() {
	win, err := NewWindow()
	if err != nil {
		log.Fatal(err)
	}
	defer win.Destroy()

	win.RunSDK()

	st := time.Now()
	for time.Since(st).Seconds() < 100 { //...
		win.Tick()
	}

}
