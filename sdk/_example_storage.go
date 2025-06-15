package main

type ExampleStruct struct {
	//<attributes>
}

func LoadExampleStruct() (*ExampleStruct, error) { //name must starts with "Load" + name of structure. Returns struct pointer and error.
	st := &ExampleStruct{} //set default values here

	return ReadJSONFile("ExampleStruct.json", st)
}

//<structures functions here>
