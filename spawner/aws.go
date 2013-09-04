package spawner

import (
  "log"
  "launchpad.net/goamz/aws"
  "launchpad.net/goamz/ec2"
  "errors"
)

type AmazonSpawnPool struct {
  // Here we will find anything necessary
  auth aws.Auth
}

func NewAmazonSpawnPool(accessKey, secretKey string) (pool AmazonSpawnPool) {
  pool = AmazonSpawnPool{auth: aws.Auth{AccessKey: accessKey, SecretKey: secretKey}}

  return
}

// The key questions around this are: we need to be able to join the nodes into etcd
// After that, we want to be able to quickly and easily scale-up and scale-down

// We want to create some stack agnostic settings here
// We'll grab the stack configuration settings from the env
// These will be sent to the environment for Fiddler

// Spawn instance will spawn an instance to whatever stack is specified in the environment
func (pool AmazonSpawnPool) Grow() (err error) {
  
  e := ec2.New(pool.auth, aws.USEast)

  options := ec2.RunInstances{
    ImageId: "ami-5b632e32", // CoreOS
    InstanceType: "t1.micro",
  }

  resp, err := e.RunInstances(&options)
  if err != nil {
    return
  }

  for _, instance := range resp.Instances {
    log.Println("Now running", instance.InstanceId)
  }

  // Next, we'll need to install Fiddler
  return
}

// Kill instance will kill a single instance in the environment
func (pool AmazonSpawnPool) Shrink() (err error) {
  // This should uh, kill one instance at random
  // Doesn't matter if it's the leader, but it might be nice to avoid that

  return errors.New("Not implemented")
}