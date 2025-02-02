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
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type Secrets struct {
	Strings map[string]string //[id]password
}

func NewSecrets() *Secrets {
	rp := &Secrets{Strings: map[string]string{}}

	//open
	js, err := os.ReadFile("secrets.json")
	if err == nil {
		json.Unmarshal(js, &rp.Strings)
	}

	return rp
}

func (rp *Secrets) Destroy() {
	//save
	js, err := json.Marshal(rp.Strings)
	if err == nil {
		os.WriteFile("secrets.json", js, 0644)
	}
}

func (rp *Secrets) Find(id string) (string, error) {
	for idd, password := range rp.Strings {
		if idd == id {
			return password, nil
		}
	}
	return "", fmt.Errorf("Password_id(%s) not found", id)
}

func (rp *Secrets) Add(password string) string {
	idBytes := make([]byte, 20) //160bit
	n, err := rand.Read(idBytes)
	if err != nil {
		log.Fatal(err)
	}
	if n != len(idBytes) {
		log.Fatal(fmt.Errorf("Invalid rand.Read() size"))
	}

	id := hex.EncodeToString(idBytes)
	rp.Strings[id] = password

	return id
}
