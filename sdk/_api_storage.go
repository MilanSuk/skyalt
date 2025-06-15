package main

func ReadJSONFile[T any](path string, defaultValues *T) (*T, error)
