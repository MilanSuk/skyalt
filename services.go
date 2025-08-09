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

import "log"

type Services struct {
	llms *LLMs
	sync *ServicesSync

	mic   *ServicesMic
	media *Media

	fnCallBuildAsync     func(ui_uid uint64, appName string, toolName string, params interface{}, fnProgress func(cmdsGob [][]byte, err error, start_time float64), fnDone func(dataJs []byte, uiGob []byte, cmdsGob []byte, err error, start_time float64)) *AppsRouterMsg
	fnGetAppPortAndTools func(appName string) (int, []*ToolsOpenAI_completion_tool, error)
}

func NewServices(media *Media) (*Services, error) {
	var err error
	srs := &Services{media: media}

	srs.mic = NewServicesMic(srs)

	srs.llms, err = NewLLMs(srs)
	if err != nil {
		return nil, err
	}

	srs.sync, err = NewServicesSync(srs)
	if err != nil {
		return nil, err
	}

	return srs, nil
}

func (srs *Services) Destroy() {
	srs.mic.Destroy()
	srs.sync.Destroy()
}

func (srs *Services) Tick(devApp_storage_changes int64) bool {
	return srs.sync.Tick(devApp_storage_changes)

}

func (srs *Services) CallBuildAsync(ui_uid uint64, appName string, toolName string, params interface{}, fnProgress func(cmdsGob [][]byte, err error, start_time float64), fnDone func(dataJs []byte, uiGob []byte, cmdsGob []byte, err error, start_time float64)) *AppsRouterMsg {
	if srs.fnCallBuildAsync == nil {
		log.Fatalf("fnCallBuildAsync is nill")
	}

	return srs.fnCallBuildAsync(ui_uid, appName, toolName, params, fnProgress, fnDone)
}
