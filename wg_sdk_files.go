package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func WgFile_read_file(path string, st any) int64 {
	js, err := os.ReadFile(path)
	if err != nil {

		if os.IsNotExist(err) {
			//create file
			return WgFile_write_file(path, st)
		}

		fmt.Println("warning: ReadFile(): ", err)
		return -1
	}

	err = json.Unmarshal(js, st)
	if err != nil {
		fmt.Println("warning: Unmarshal(): ", err)
		return -1
	}

	fmt.Println("File open:", path)
	return OsFileTime(path)
}

func WgFile_write_file(path string, st any) int64 {

	js, err := json.MarshalIndent(st, "", "")
	if err != nil {
		fmt.Println("warning: MarshalIndent(): ", err)
	}

	origJs, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(js, origJs) {
		os.MkdirAll(filepath.Dir(path), os.ModePerm)

		err = os.WriteFile(path, js, 0644)
		if err != nil {
			fmt.Println("warning: WriteFile(): ", err)
		}
		fmt.Println("File saved:", path)
	}

	return OsFileTime(path)
}

type WgFile struct {
	data        interface{}
	time_stamps int64
}

type WgFiles struct {
	files map[string]*WgFile
	lock  sync.Mutex
}

var g_files WgFiles

func WgFiles_getPath(name string) string {
	return filepath.Join("disk", name)
}

func WgFiles_Load[T any](name string, def *T) *T {
	path := WgFiles_getPath(name)

	g_files.lock.Lock()
	defer g_files.lock.Unlock()

	if g_files.files == nil {
		g_files.files = make(map[string]*WgFile)
	}

	//find
	st, found := g_files.files[path]
	if found {
		stt, ok := st.data.(*T)
		if !ok {
			fmt.Printf("Runtime error: bad casting for path(%s)", path)
		}
		return stt
	}

	//add
	time_stamps := WgFile_read_file(path+".json", def)
	if time_stamps > 0 {
		g_files.files[path] = &WgFile{data: def, time_stamps: time_stamps}
	}
	return def
}

func WgFiles_Delete(name string) {
	path := WgFiles_getPath(name)

	g_files.lock.Lock()
	defer g_files.lock.Unlock()

	os.Remove(path + ".json")
	delete(g_files.files, path)
}

func WgFiles_Save() {
	g_files.lock.Lock()
	defer g_files.lock.Unlock()

	for path, it := range g_files.files {
		g_files.files[path].time_stamps = WgFile_write_file(path+".json", it.data)
	}
	//g_files.files = nil	//some threads/jobs can have old pointer
}

func WgFiles_Refresh() {
	g_files.lock.Lock()
	defer g_files.lock.Unlock()

	for path, it := range g_files.files {
		if it.time_stamps != OsFileTime(path+".json") {
			delete(g_files.files, path) //remove, so it's re-created from file
		}
	}

}

func OpenFile_OsmSettings() *Osm {
	return WgFiles_Load("osm_settings", &Osm{Enable: true, Tiles_url: "https://tile.openstreetmap.org/{z}/{x}/{y}.png", Cache_path: "disk/osm_tiles_cache.sqlite", Copyright: "(c)OpenStreetMap contributors", Copyright_url: "https://www.openstreetmap.org/copyright"})
}

func OpenFile_Microphone() *Microphone {
	return WgFiles_Load("microphone", &Microphone{Enable: true, SampleRate: 44100, Channels: 1})
}

func OpenFile_DeviceSettings() *DeviceSettings {
	st := WgFiles_Load("device_settings", &DeviceSettings{})
	st.Check()
	return st
}

/*func SaveFile_Settings() {
	WgFile_write_file(WgFiles_getPath("device_settings"), OpenFile_Settings())
}*/

func OpenFile_Root() *Root {
	return WgFiles_Load("root", &Root{})
}

func OpenFile_Chat(path string) *Agent {
	return WgFiles_Load(path, &Agent{})

}

func RemoveFile_Chat(path string) {
	WgFiles_Delete(path)
}

func OpenDir_llms_logins() ([]string, error) {
	dir, err := os.ReadDir("disk/llms_logins")
	if err != nil {
		return nil, err
	}

	var list []string
	for _, it := range dir {
		nm := strings.TrimSuffix(it.Name(), filepath.Ext(it.Name())) //name without .json
		list = append(list, filepath.Join("llms_logins", nm))
	}

	return list, nil
}
func OpenFile_LLMLogin(path string) *LLMLogin {
	return WgFiles_Load(path, &LLMLogin{})
}

func FindLoginChatModel(chat_model string) (*LLMLogin, string) {
	logins, err := OpenDir_llms_logins()
	if err != nil {
		return nil, "" //err ....
	}

	for _, login_path := range logins {
		login := OpenFile_LLMLogin(login_path)
		for _, m := range login.ChatModels {
			if m.Name == chat_model {
				return login, login_path
			}
		}
	}
	return nil, ""
}

func OpenDir_agents_properties() ([]string, error) {
	dir, err := os.ReadDir("disk/agents_properties")
	if err != nil {
		return nil, err
	}

	var list []string
	for _, it := range dir {
		nm := strings.TrimSuffix(it.Name(), filepath.Ext(it.Name())) //name without .json
		list = append(list, filepath.Join("agents_properties", nm))
	}

	return list, nil
}
func OpenFile_Agent_properties(path string) *Agent_properties {
	return WgFiles_Load(path, &Agent_properties{})
}

//ui_slider, ui_settings - user must confirm and then it's written ......
//- change again = re-run chat from that point
