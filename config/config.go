package config

import (
  "io/ioutil"
  "encoding/json"
)

type DockerConf struct {
  Container string
  Args []string
}

type FiddlerConf struct {
  Docker *DockerConf
}

func LoadFiddlerConfig(c string) (conf *FiddlerConf, err error) {
  // First, let's try opening and reading the file
  lines, err := ioutil.ReadFile(c) // For read access.
  if err != nil {
    return nil, err
  }

  // Now, let's deserialize the JSON
  err = json.Unmarshal([]byte(lines), &conf)
  if err != nil {
    return nil, err
  }
  
  return // Conf is set
}