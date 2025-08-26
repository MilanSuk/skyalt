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
	"errors"
)

func _ToolsCaller_UpdateDev(port int) error {
	cl, err := NewToolsClient("localhost", port)
	if err != nil {
		return err
	}

	defer cl.Destroy()

	//send
	err = cl.WriteArray([]byte("update_dev"))
	if err != nil {
		return err
	}
	return nil
}

func _ToolsCaller_CallBuild(port int, msg_id uint64, ui_uid uint64, toolName string, paramsJs []byte) ([]byte, []byte, []byte, error) {
	cl, err := NewToolsClient("localhost", port)
	if err != nil {
		return nil, nil, nil, err
	}
	defer cl.Destroy()

	//send
	err = cl.WriteArray([]byte("build"))
	if err != nil {
		return nil, nil, nil, err
	}

	err = cl.WriteInt(msg_id) //msg_id
	if err != nil {
		return nil, nil, nil, err
	}

	err = cl.WriteInt(ui_uid) //UI UID
	if err != nil {
		return nil, nil, nil, err
	}
	err = cl.WriteArray([]byte(toolName)) //function name
	if err != nil {
		return nil, nil, nil, err
	}
	err = cl.WriteArray(paramsJs) //params
	if err != nil {
		return nil, nil, nil, err
	}

	errStr, err := cl.ReadArray() //output error
	if err != nil {
		return nil, nil, nil, err
	}
	out_dataJs, err := cl.ReadArray() //output data
	if err != nil {
		return nil, nil, nil, err
	}
	out_uiGob, err := cl.ReadArray() //output UI
	if err != nil {
		return nil, nil, nil, err
	}
	out_cmdsBog, err := cl.ReadArray() //output cmds
	if err != nil {
		return nil, nil, nil, err
	}

	var out_err error
	if len(errStr) > 0 {
		out_err = errors.New(string(errStr))
		LogsError(out_err)
	}

	return out_dataJs, out_uiGob, out_cmdsBog, out_err
}

func _ToolsCaller_CallChange(port int, msg_id uint64, ui_uid uint64, change ToolsSdkChange) ([]byte, []byte, error) {
	cl, err := NewToolsClient("localhost", port)
	if err != nil {
		return nil, nil, err
	}

	defer cl.Destroy()

	//send
	err = cl.WriteArray([]byte("change"))
	if err != nil {
		return nil, nil, err
	}
	err = cl.WriteInt(msg_id)
	if err != nil {
		return nil, nil, err
	}
	err = cl.WriteInt(ui_uid)
	if err != nil {
		return nil, nil, err
	}
	changeJs, _ := LogsJsonMarshal(change)

	err = cl.WriteArray(changeJs)
	if err != nil {
		return nil, nil, err
	}

	errStr, err := cl.ReadArray()
	if err != nil {
		return nil, nil, err
	}
	dataJs, err := cl.ReadArray()
	if err != nil {
		return nil, nil, err
	}
	cmdsGob, _ := cl.ReadArray()

	if len(errStr) > 0 {
		err := errors.New(string(errStr))
		LogsError(err)
		return nil, nil, err
	}

	return dataJs, cmdsGob, nil

}

func _ToolsCaller_CallUpdate(port int, msg_id uint64, ui_uid uint64, sub_uid uint64) ([]byte, []byte, error) {
	cl, err := NewToolsClient("localhost", port)
	if err != nil {
		return nil, nil, err
	}
	defer cl.Destroy()

	//send
	err = cl.WriteArray([]byte("update"))
	if err != nil {
		return nil, nil, err
	}
	err = cl.WriteInt(msg_id)
	if err != nil {
		return nil, nil, err
	}
	err = cl.WriteInt(ui_uid)
	if err != nil {
		return nil, nil, err
	}
	err = cl.WriteInt(sub_uid)
	if err != nil {
		return nil, nil, err
	}

	errStr, err := cl.ReadArray()
	if err != nil {
		return nil, nil, err
	}
	subUiGob, err := cl.ReadArray()
	if err != nil {
		return nil, nil, err
	}
	cmdsGob, _ := cl.ReadArray()

	if len(errStr) > 0 {
		err := errors.New(string(errStr))
		LogsError(err)
		return nil, nil, err
	}

	return subUiGob, cmdsGob, nil
}
