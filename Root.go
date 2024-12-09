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

func (st *Root) Build() {
	st.lock.Lock()
	defer st.lock.Unlock()

	st.layout.SetColumn(0, 1, 100)
	//st.layout.SetColumn(1, 1, 15)
	//st.layout.SetColumn(2, 1, 15)
	st.layout.SetRow(0, 1, 100)

	st.layout.AddEnv(0, 0, 1, 1, NewFile_Env())
	//st.layout.AddCounter(1, 0, 1, 1, NewFile_Counter())
	//st.layout.AddLogs(2, 0, 1, 1, NewFile_Logs())
}
