package main

import (
	"encoding/json"
	"log"
)

func _callGlobalInits() {
	var err error

	err = GetMapTile_global_init()
	if err != nil {
		log.Fatal(err)
	}
}
func _callGlobalDestroys() {
	var err error

	defer func() {
		err = GetMapTile_global_destroy()
		if err != nil {
			log.Fatal(err)
		}
	}()
}
func FindToolRunFunc(funcName string, jsParams []byte) (func(caller *ToolCaller, ui *UI) error, interface{}) {
	switch funcName {
	case "AddEvent":
		st := AddEvent{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "AddEventGroup":
		st := AddEventGroup{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "Calculate":
		st := Calculate{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "ChangeEventDate":
		st := ChangeEventDate{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "ChangeEventGroup":
		st := ChangeEventGroup{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "DeleteActivity":
		st := DeleteActivity{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "DeleteEmailLogin":
		st := DeleteEmailLogin{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "DeleteEvent":
		st := DeleteEvent{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "DownloadFile":
		st := DownloadFile{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "ExportActivitiesIntoFolder":
		st := ExportActivitiesIntoFolder{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "ExportActivityIntoGPX":
		st := ExportActivityIntoGPX{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "ExportEventsIntoICS":
		st := ExportEventsIntoICS{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "GetDeviceSettings":
		st := GetDeviceSettings{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "GetEmailLogins":
		st := GetEmailLogins{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "GetEventsInDays":
		st := GetEventsInDays{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "GetListOfActivities":
		st := GetListOfActivities{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "GetListOfGroups":
		st := GetListOfGroups{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "GetMapSettings":
		st := GetMapSettings{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "GetMapTile":
		st := GetMapTile{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "GetWeekDays":
		st := GetWeekDays{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "ImportGPXFolderToActivities":
		st := ImportGPXFolderToActivities{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "ImportGPXToActivities":
		st := ImportGPXToActivities{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "ImportICSToEvents":
		st := ImportICSToEvents{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "LLMxAICompleteChat":
		st := LLMxAICompleteChat{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "LLMxAIGenerateImage":
		st := LLMxAIGenerateImage{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "RecordMicrophone":
		st := RecordMicrophone{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "SendEmail":
		st := SendEmail{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "SetDeviceDPI":
		st := SetDeviceDPI{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "SetDeviceDPIDefault":
		st := SetDeviceDPIDefault{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "SetDeviceFullscreen":
		st := SetDeviceFullscreen{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "SetDeviceStats":
		st := SetDeviceStats{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "ShowActivity":
		st := ShowActivity{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "ShowActivityElevationChart":
		st := ShowActivityElevationChart{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "ShowActivityMap":
		st := ShowActivityMap{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "ShowActivityPacesChart":
		st := ShowActivityPacesChart{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "ShowActivityStatistic":
		st := ShowActivityStatistic{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "ShowAddEmailLogin":
		st := ShowAddEmailLogin{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "ShowChat":
		st := ShowChat{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "ShowDayCalendar":
		st := ShowDayCalendar{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "ShowDeviceSettings":
		st := ShowDeviceSettings{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "ShowEmailLogin":
		st := ShowEmailLogin{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "ShowEvent":
		st := ShowEvent{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "ShowGroups":
		st := ShowGroups{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "ShowLLMWhispercppSettings":
		st := ShowLLMWhispercppSettings{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "ShowLLMxAISettings":
		st := ShowLLMxAISettings{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "ShowListOfActivities":
		st := ShowListOfActivities{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "ShowListOfEmails":
		st := ShowListOfEmails{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "ShowMapSettings":
		st := ShowMapSettings{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "ShowMonthCalendar":
		st := ShowMonthCalendar{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "ShowRoot":
		st := ShowRoot{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "ShowUserBodyMeasurements":
		st := ShowUserBodyMeasurements{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "ShowYearCalendar":
		st := ShowYearCalendar{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st
	case "WhispercppTranscribe":
		st := WhispercppTranscribe{}
		err := json.Unmarshal(jsParams, &st)
		if err != nil {
			return nil, nil
		}
		return st.run, &st

	}
	return nil, nil
}
