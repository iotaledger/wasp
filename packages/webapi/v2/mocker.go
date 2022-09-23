package v2

import (
	"embed"
	"encoding/json"
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

func getStructName(i interface{}) string {
	t := reflect.TypeOf(i)

	if t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	}

	return t.Name()
}

func (m *Mocker) AddModel(i interface{}) {
	name := getStructName(i)

	if jsonMockData, ok := m.mockData[name]; ok {
		err := json.Unmarshal(jsonMockData, &i)
		if err != nil {
			log.Fatal(err)
		}
	}

	m.mockedModels[name] = i
}

func (m *Mocker) LoadMockFiles() {
	files, err := mockDir.ReadDir("models/mock")
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		mockData, err := mockDir.ReadFile("models/mock/" + file.Name())
		if err != nil {
			panic(err.Error())
		}

		name := strings.Replace(file.Name(), ".json", "", 1)
		m.mockData[name] = mockData
	}
}

func (m *Mocker) GetMockedStruct(i interface{}) interface{} {
	name := getStructName(i)

	if mockedStruct, ok := m.mockedModels[name]; ok {
		return mockedStruct
	}

	return nil
}
