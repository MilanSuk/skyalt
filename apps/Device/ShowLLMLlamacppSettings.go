package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

// Show Llama.cpp settings.
type ShowLLMLlamacppSettings struct {
}

func (st *ShowLLMLlamacppSettings) run(caller *ToolCaller, ui *UI) error {
	source_llama, err := NewLLMLlamacpp("")
	if err != nil {
		return err
	}

	source_llama.Check()

	ui.SetColumn(0, 1, 5)
	ui.SetColumn(1, 1, 20)

	ui.AddTextLabel(0, 0, 2, 1, "Llama.cpp")

	y := 1

	wChanged := func() error {
		source_llama.Check() //reload models
		return nil
	}

	ui.AddText(0, y, 1, 1, "Address : port")
	AddrDiv := ui.AddLayout(1, y, 1, 1)
	{
		AddrDiv.SetColumn(0, 1, 100)
		AddrDiv.SetColumn(1, 0.5, 0.5)
		AddrDiv.SetColumn(2, 1, 4)
		AddrDiv.SetColumn(3, 1, 4)

		ad := AddrDiv.AddEditboxString(0, 0, 1, 1, &source_llama.Address)
		ad.changed = wChanged

		AddrDiv.AddText(1, 0, 1, 1, ":")

		pt := AddrDiv.AddEditboxInt(2, 0, 1, 1, &source_llama.Port)
		pt.changed = wChanged

		TestOKDia := ui.AddDialog("test_ok")
		TestOKDia.UI.SetColumn(0, 5, 7)
		tx := TestOKDia.UI.AddText(0, 0, 1, 1, "OK - Server is running!")
		tx.Align_h = 1

		TestErrDia := ui.AddDialog("test_err")
		TestErrDia.UI.SetColumn(0, 5, 5)
		TestErrDia.UI.Border_cd = UI_GetPalette().E
		tx = TestErrDia.UI.AddText(0, 0, 1, 1, "Error - Server not found")
		tx.Align_h = 1
		tx.Cd = UI_GetPalette().E

		TestBt := AddrDiv.AddButton(3, 0, 1, 1, "Test")
		TestBt.clicked = func() error {
			status, err := st.SetModel("", source_llama)
			if err == nil && status == 200 {
				TestOKDia.OpenRelative(TestBt.layout, caller)
			} else {
				TestErrDia.OpenRelative(TestBt.layout, caller)
			}
			return nil
		}
	}
	y++

	ui.AddText(0, y, 1, 1, "Command example")
	ui.AddText(1, y, 1, 1, fmt.Sprintf("./llama-server --port %d -m models/llama-3.2-1b-instruct-q8_0.gguf", source_llama.Port))
	y++

	return nil
}

func (llama *ShowLLMLlamacppSettings) SetModel(model string, source *LLMLlamacpp) (int, error) {
	source.lock.Lock()
	defer source.lock.Unlock()

	body := bytes.NewReader(nil)
	req, err := http.NewRequest(http.MethodGet, source.GetUrlHealth(), body)
	if err != nil {
		return -1, fmt.Errorf("NewRequest() failed: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")

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
