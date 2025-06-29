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
	"encoding/json"
	"errors"
	"fmt"
)

func _ToolsCaller_UpdateDev(port int, fnLog func(err error) error) error {
	cl, err := NewToolsClient("localhost", port)
	if fnLog(err) == nil {
		defer cl.Destroy()

		//send
		err = cl.WriteArray([]byte("update_dev"))
		fnLog(err)
	}

	return fmt.Errorf("connection failed")
}

func _ToolsCaller_CallBuild(port int, msg_id uint64, ui_uid uint64, funcName string, params []byte, fnLog func(err error) error) ([]byte, []byte, []byte, error) {
	cl, err := NewToolsClient("localhost", port)
	if fnLog(err) == nil {
		defer cl.Destroy()

		//send
		err = cl.WriteArray([]byte("build"))
		if fnLog(err) == nil {
			err = cl.WriteInt(msg_id) //msg_id
			if fnLog(err) == nil {

				err = cl.WriteInt(ui_uid) //UI UID
				if fnLog(err) == nil {
					err = cl.WriteArray([]byte(funcName)) //function name
					if fnLog(err) == nil {
						err = cl.WriteArray(params) //params
						if fnLog(err) == nil {

							errStr, err := cl.ReadArray() //output error
							if fnLog(err) == nil {
								out_dataJs, err := cl.ReadArray() //output data
								if fnLog(err) == nil {
									out_uiJs, err := cl.ReadArray() //output UI
									if fnLog(err) == nil {
										out_cmdsJs, err := cl.ReadArray() //output cmds
										if fnLog(err) == nil {

											var out_err error
											if len(errStr) > 0 {
												out_err = errors.New(string(errStr))
											}

											return out_dataJs, out_uiJs, out_cmdsJs, out_err
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return nil, nil, nil, fmt.Errorf("connection failed")
}

func _ToolsCaller_CallChange(port int, msg_id uint64, ui_uid uint64, change ToolsSdkChange, fnLog func(err error) error) ([]byte, []byte, error) {
	cl, err := NewToolsClient("localhost", port)
	if fnLog(err) == nil {
		defer cl.Destroy()

		//send
		err = cl.WriteArray([]byte("change"))
		if fnLog(err) == nil {
			err = cl.WriteInt(msg_id)
			if fnLog(err) == nil {
				err = cl.WriteInt(ui_uid)
				if fnLog(err) == nil {
					changeJs, err := json.Marshal(change)
					if fnLog(err) == nil {
						err = cl.WriteArray(changeJs)
						if fnLog(err) == nil {

							errStr, err := cl.ReadArray()
							if fnLog(err) == nil {
								dataJs, err := cl.ReadArray()
								if fnLog(err) == nil {
									cmdsJs, err := cl.ReadArray()
									fnLog(err)

									if len(errStr) > 0 {
										return nil, nil, errors.New(string(errStr))
									}

									return dataJs, cmdsJs, nil
								}
							}
						}
					}
				}
			}
		}
	}

	return nil, nil, fmt.Errorf("connection failed")
}

func _ToolsCaller_CallUpdate(port int, msg_id uint64, ui_uid uint64, sub_uid uint64, fnLog func(err error) error) ([]byte, []byte, error) {
	cl, err := NewToolsClient("localhost", port)
	if fnLog(err) == nil {
		defer cl.Destroy()

		//send
		err = cl.WriteArray([]byte("update"))
		if fnLog(err) == nil {
			err = cl.WriteInt(msg_id)
			if fnLog(err) == nil {
				err = cl.WriteInt(ui_uid)
				if fnLog(err) == nil {
					err = cl.WriteInt(sub_uid)
					if fnLog(err) == nil {

						errStr, err := cl.ReadArray()
						if fnLog(err) == nil {
							subUiJs, err := cl.ReadArray()
							if fnLog(err) == nil {
								cmdsJs, err := cl.ReadArray()
								fnLog(err)

								if len(errStr) > 0 {
									return nil, nil, errors.New(string(errStr))
								}

								return subUiJs, cmdsJs, nil
							}
						}
					}
				}
			}
		}
	}

	return nil, nil, fmt.Errorf("connection failed")
}
