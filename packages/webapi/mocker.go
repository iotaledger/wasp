package webapi

import (
	"embed"
	"log"
	"reflect"
	"strings"
)

//go:embed models/mock
var mockDir embed.FS

type Mocker struct {
	mockData     map[string][]byte
	mockedModels map[string]interface{}
}

func NewMocker() *Mocker {
	return &Mocker{
		mockData:     map[string][]byte{},
		mockedModels: map[string]interface{}{},
	}
}

func createNewInstance(i interface{}) reflect.Value {
	t := reflect.TypeOf(i)

	if t.Kind() == reflect.Array || t.Kind() == reflect.Slice {
		newSlice := reflect.MakeSlice(t, 0, 0)
		return reflect.New(newSlice.Type())
	}

	if t.Kind() == reflect.Ptr {
		return reflect.New(t.Elem())
	}

	return reflect.New(t)
}

func (m *Mocker) LoadMockFiles() {
	files, err := mockDir.ReadDir("models/mock")
	if err != nil {
		log.Fatalf("Mocker: %v", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		mockData, err := mockDir.ReadFile("models/mock/" + file.Name())
		if err != nil {
			log.Fatalf("Mocker[%s] %v", file.Name(), err.Error())
		}

		name := strings.Replace(file.Name(), ".json", "", 1)
		m.mockData[name] = mockData
	}
}

func (m *Mocker) Get(i interface{}) interface{} {
	model := createNewInstance(i)
	instance := model.Interface()
	// For now just return the empty instance
	return instance

	/*
		name := getStructName(i)

		if jsonMockData, ok := m.mockData[name]; ok {
			err := json.Unmarshal(jsonMockData, instance)
			if err != nil {
				log.Fatalf("Mocker [%s] %v", name, err)
			}
		}

		return instance*/
}
