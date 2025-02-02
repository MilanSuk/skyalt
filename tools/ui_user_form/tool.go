package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

// Show a input form on screen. It returns JSON with values from user.
type ui_user_form struct {
	Description string //What labels, types, and descriptions you wanna get from the user.
}

type Item struct {
	Label       string
	Type        string
	Description string

	Value string
}

func (st *ui_user_form) run() string {

	SystemPrompt := "You are an AI assistant, who enjoys precision and carefully follows the user's requirements. You answer in JSON format."

	UserPrompt := "Here is the prompt from the user:\n"
	UserPrompt += st.Description + "\n\n"
	UserPrompt += `Your job is to convert the prompt to JSON array with format [{"label": "<username>", "type": "<string,integer,float>", "description": "<How do you wanna be called>"}]. Output only JSON, no explanation.`

	fmt.Println("UserPrompt:", UserPrompt)

	js := SDK_RunAgent("main", 20, 20000, SystemPrompt, UserPrompt)

	fmt.Println("Js:", string(js))

	var ok bool
	js, ok = strings.CutPrefix(js, "```json")
	if !ok {
		js, _ = strings.CutPrefix(js, "```")
	}
	js, _ = strings.CutSuffix(js, "```") //end

	//parse
	var items []Item
	err := json.Unmarshal([]byte(js), &items)
	if err != nil {
		log.Fatal(err)
	}

	//render ....
	for i, it := range items {
		fmt.Println(it.Label, "//", it.Description)
		var input string
		//fmt.Scanln(&input)
		input = "786123"
		items[i].Value = input
	}

	//convert back
	answerJs, err := json.Marshal(items)
	if err != nil {
		log.Fatal(err)
	}

	//wait when user click Confirm ....

	return string(answerJs)
}
