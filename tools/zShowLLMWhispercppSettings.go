package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

// Show Whisper.cpp settings.
type ShowLLMWhispercppSettings struct {
}

func (st *ShowLLMWhispercppSettings) run(caller *ToolCaller, ui *UI) error {
	source_wsp, err := NewLLMWhispercpp_wsp("", caller)
	if err != nil {
		return err
	}

	source_wsp.Check()

	ui.SetColumn(0, 1, 5)
	ui.SetColumn(1, 1, 20)

	ui.AddTextLabel(0, 0, 2, 1, "Whisper.cpp")

	y := 1

	wChanged := func() error {
		source_wsp.Check() //reload models
		return nil
	}

	ui.AddText(0, y, 1, 1, "Folder")
	fd := ui.AddFilePickerButton(1, y, 1, 1, &source_wsp.Folder, true, true)
	fd.changed = wChanged
	y++

	ui.AddText(0, y, 1, 1, "Address")
	AddrDiv := ui.AddLayout(1, y, 1, 1)
	{
		AddrDiv.SetColumn(0, 1, 100)
		AddrDiv.SetColumn(1, 1, 4)
		AddrDiv.SetColumn(2, 1, 4)
		AddrDiv.SetColumn(3, 0, 4)

		ad := AddrDiv.AddEditboxString(0, 0, 1, 1, &source_wsp.Address)
		ad.changed = wChanged
		if source_wsp.Address == "" {
			ad.Error = "Empty"
		}
		pt := AddrDiv.AddEditboxInt(1, 0, 1, 1, &source_wsp.Port)
		pt.changed = wChanged

		TestOKDia := ui.AddDialog("test_ok")
		TestOKDia.UI.SetColumn(0, 5, 7)
		tx := TestOKDia.UI.AddText(0, 0, 1, 1, "OK - server is running")
		tx.Align_h = 1

		TestErrDia := ui.AddDialog("test_err")
		TestErrDia.UI.SetColumn(0, 5, 5)
		tx = TestErrDia.UI.AddText(0, 0, 1, 1, "Not found")
		tx.Align_h = 1
		tx.Cd = caller.GetPalette().E

		TestBt := AddrDiv.AddButton(2, 0, 1, 1, "Test")
		TestBt.clicked = func() error {
			status, err := st.SetModel(source_wsp.ModelName, source_wsp)
			if err == nil && status == 200 {
				TestOKDia.OpenCentered(caller)
			} else {
				TestErrDia.OpenCentered(caller)
			}
			return nil
		}
	}
	y++

	//Models
	ui.SetRowFromSub(y, 1, 100)
	ModelsDiv := ui.AddLayout(0, y, 2, 1)
	y++
	ModelsDiv.SetColumn(0, 5, 5)
	ModelsDiv.SetColumn(1, 1, 100)
	ModelsDiv.SetColumn(2, 1, 100)
	my := 0

	ModelsDiv.AddText(0, my, 2, 1, "Text to Speech")
	for _, it := range source_wsp.Models {
		ModelsDiv.AddText(1, my, 1, 1, it.Label)
		my++
	}

	btReload := ui.AddButton(1, y, 1, 1, "Reload list")
	btReload.clicked = func() error {
		return source_wsp.ReloadModels()
	}

	return nil
}

func (wsp *ShowLLMWhispercppSettings) SetModel(model string, source *LLMWhispercpp) (int, error) {
	source.lock.Lock()
	defer source.lock.Unlock()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("model", source.getModelPath(model))
	writer.Close()

	req, err := http.NewRequest(http.MethodPost, source.GetUrlLoadModel(), body)
	if err != nil {
		return -1, fmt.Errorf("NewRequest() failed: %w", err)
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return -1, fmt.Errorf("Do() failed: %w", err)
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return res.StatusCode, fmt.Errorf("ReadAll() failed: %w", err)
	}

	if res.StatusCode != 200 {
		return res.StatusCode, fmt.Errorf("statusCode != 200, response: %s", resBody)
	}

	return res.StatusCode, nil
}
