package spawner

// TODO: Make platform agnostic
// TODO: Take keys not from ENV?
// TODO: Find a way to get Fiddler script to run
// Are we going to have to run an SSH to run the install?

import (
  "fmt"
  "os"
  "errors"
)

type SpawnPool interface {
  Grow() (err error)
  Shrink() (err error)
}

// GetSpawnPool will return spawn pool from the environment or etcd
func GetSpawnPool() (pool SpawnPool, err error) {
  // First, we'll try to grab the environment from env
  if len(os.Getenv("cloud")) == 0 {
    return nil, errors.New("Must provide cloud env variable to run spawner")
  }

  // We have a cloud environment variable
  switch os.Getenv("cloud") {
  case "aws":
    var key, secret string
    key, secret = os.Getenv("key"), os.Getenv("secret")
    if len(key) == 0 || len(secret) == 0 {
      return nil, errors.New("Must provide key and secret to spawn AWS instances")
    }
    return NewAmazonSpawnPool(key, secret), nil
  default:
    return nil, errors.New(fmt.Sprintf("Unknown cloud: %s", os.Getenv("cloud")))
  }
}