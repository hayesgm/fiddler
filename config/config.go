package config

import (
  "io/ioutil"
  "encoding/json"
  "strings"
  "net/http"
)

type DockerConf struct {
  Container string
  Run string
  Args []string
}

type ScaleConf struct {
  Min int
  Max int
  Grow map[string]string
  Shrink map[string]string
}

type FiddlerConf struct {
  Docker *DockerConf
  Scale *ScaleConf
}

func LoadFiddlerConfig(c string) (conf *FiddlerConf, err error) {
  var lines []byte

  if strings.HasPrefix(c,"http") {
    resp, err := http.Get(c)
    if err != nil {
      return nil, err
    }
    defer resp.Body.Close()

    lines, err = ioutil.ReadAll(resp.Body)
    if err != nil {
      return nil, err
    }
  } else {
    // First, let's try opening and reading the file
    lines, err = ioutil.ReadFile(c) // For read access.
    if err != nil {
      return nil, err
    }
  }

  // Now, let's deserialize the JSON
  err = json.Unmarshal([]byte(lines), &conf)
  if err != nil {
    return nil, err
  }
  
  return // Conf is set
}