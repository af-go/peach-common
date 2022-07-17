package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/go-logr/logr"
	"gopkg.in/yaml.v3"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

// LoadYaml load properties from yaml and convert to dot properties, then set into map
func LoadYaml(content []byte, log logr.Logger) (map[string]string, error) {
	m := make(map[string]interface{})
	err := yaml.Unmarshal(content, &m)
	if err != nil {
		return nil, err
	}
	converted := make(map[string]string)
	for k, v := range m {
		flatValue(v, k, converted)
	}
	return converted, nil
}

func flatValue(v interface{}, parent string, m map[string]string) {
	typ := reflect.TypeOf(v).Kind()
	if typ == reflect.Map {
		for k, sv := range v.(map[interface{}]interface{}) {
			key := fmt.Sprintf("%s", k)
			if parent != "" {
				key = fmt.Sprintf("%s.%s", parent, k)
			}
			flatValue(sv, key, m)
		}
	} else if typ == reflect.Int {
		value := fmt.Sprintf("%v", v)
		m[parent] = value
	} else if typ == reflect.String {
		value := fmt.Sprintf("%v", v)
		m[parent] = value
	} else if typ == reflect.Slice {
		for i, sv := range v.([]interface{}) {
			key := fmt.Sprintf("[%d]", i)
			if parent != "" {
				key = fmt.Sprintf("%s[%d]", parent, i)
			}
			flatValue(sv, key, m)
		}
	}
}

// Load load config file and unmarshall
func Load(filename string, v interface{}, log logr.Logger) error {
	content, err := ioutil.ReadFile(filepath.Clean(filename)) // fix gosec G304
	if err != nil {
		log.Error(err, "failed to open file", "filename", filename)
		return err
	}
	if strings.HasSuffix(filename, ".json") {
		err = json.Unmarshal(content, v)
		if err != nil {
			log.Error(err, "failed to unmarshall file", "filename", filename)
			return err
		}
	} else if strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml") {
		err = yaml.Unmarshal(content, v)
		if err != nil {
			log.Error(err, "failed to unmarshall file", "filename", filename)
			return err
		}
	} else {
		log.Info("unsupported file type, neither json, nor yaml", "filename", filename)
	}
	return err
}

func NewLogger(enableDebug bool) logr.Logger {
	logger := zap.New(func(o *zap.Options) {
		o.Development = enableDebug
	})
	return logger
}
