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
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"fmt"
	"os"
	"strings"
)

type ToolsSecret struct {
	Alias string
	Value string
}

type ToolsSecrets struct {
	path  string
	items []ToolsSecret //[alias]
	key   string
}

func NewToolsSecrets(path string) (*ToolsSecrets, error) {
	sec := &ToolsSecrets{path: path, key: "skyalt"}

	//open
	cipher, err := os.ReadFile(path)
	if err == nil {
		plain, err := sec.decryptAESGCM(cipher)
		if LogsError(err) != nil {
			return nil, err
		}

		lines := strings.Split(string(plain), "\n")
		for _, ln := range lines {
			ln = strings.TrimSpace(ln)
			if ln == "" {
				continue //skip empty
			}

			d := strings.IndexAny(ln, " \t")
			if d >= 0 {
				sec.items = append(sec.items, ToolsSecret{Alias: strings.TrimSpace(ln[:d]), Value: strings.TrimSpace(ln[d:])})
			} else {
				fmt.Printf("Warning: Secret line '%s' has no separator\n", ln)
			}
		}
	}

	return sec, nil
}

func (sec *ToolsSecrets) Destroy() error {
	return nil
}

func (sec *ToolsSecrets) ReplaceAliases(code string) string {
	for _, it := range sec.items {
		code = strings.ReplaceAll(code, fmt.Sprintf(`"%s"`, it.Alias), fmt.Sprintf(`SdkGetSecret("%s")`, it.Alias))
	}
	return code
}

//Same func in sdk.go
/*func (sec *ToolsSecrets) encryptAESGCM(plainText []byte) ([]byte, error) {
	key := sha256.Sum256([]byte(sec.key))

	block, err := aes.NewCipher(key[:])
	if LogsError(err) != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if LogsError(err) != nil {
		return nil, err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if LogsError(err) != nil {
		return nil, err
	}

	ciphertext := aesGCM.Seal(nil, nonce, plainText, nil)
	return append(nonce, ciphertext...), nil
}*/

func (sec *ToolsSecrets) decryptAESGCM(cipherText []byte) ([]byte, error) {
	key := sha256.Sum256([]byte(sec.key))

	block, err := aes.NewCipher(key[:])
	if LogsError(err) != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if LogsError(err) != nil {
		return nil, err
	}

	nonceSize := aesGCM.NonceSize()
	if len(cipherText) < nonceSize {
		return nil, LogsErrorf("cipherText too short")
	}

	nonce, ciphertext := cipherText[:nonceSize], cipherText[nonceSize:]
	plainText, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if LogsError(err) != nil {
		return nil, err
	}

	return plainText, nil
}
