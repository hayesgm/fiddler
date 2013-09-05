package spawner

import (
  "log"
  "launchpad.net/goamz/aws"
  "launchpad.net/goamz/ec2"
  "errors"
  "fmt"
)

type AmazonSpawnPool struct {
  // Here we will find anything necessary
  auth aws.Auth
  zone aws.Region
  name string
  privateKey string
  sg ec2.SecurityGroup
}

func NewAmazonSpawnPool(name, accessKey, secretKey, zone, privateKey string) (pool *AmazonSpawnPool, err error) {
  var awsZone, ok = aws.Regions[zone]
  if !ok {
    return nil, errors.New(fmt.Sprintf("Unable to find zone: %s", zone))
  }

  pool = &AmazonSpawnPool{name: name, auth: aws.Auth{AccessKey: accessKey, SecretKey: secretKey}, zone: awsZone, privateKey: privateKey}
  
  pool.sg, err = pool.findOrCreateSecurityGroup()
  if err != nil {
    return
  }

  return
}

// The key questions around this are: we need to be able to join the nodes into etcd
// After that, we want to be able to quickly and easily scale-up and scale-down

// We want to create some stack agnostic settings here
// We'll grab the stack configuration settings from the env
// These will be sent to the environment for Fiddler

// Spawn instance will spawn an instance to whatever stack is specified in the environment
func (pool AmazonSpawnPool) Grow(conf string) (err error) {
  
  e := ec2.New(pool.auth, aws.USEast)

  options := ec2.RunInstances{
    ImageId: "ami-5b632e32", // CoreOS
    InstanceType: "t1.micro",
    KeyName: "fiddler", // TODO: We should add this
    SecurityGroups: []ec2.SecurityGroup{pool.sg},
    UserData: []byte(fmt.Sprintf("https://fiddler/%s/%s", pool.sg.Id, pool.name)), // We need something consistent here
  }

  resp, err := e.RunInstances(&options)
  if err != nil {
    return
  }

  for _, instance := range resp.Instances {
    log.Println("Now running", instance.InstanceId)
    // Next, we'll need to install Fiddler
    // It's kind of annoying, but we can ssh into our new box
    installCmd := fmt.Sprintf("\\curl -L https://raw.github.com/hayesgm/fiddler/master/install.sh | bash -s %s", conf)
    err = sshCommand(instance.IPAddress, "core", pool.privateKey, installCmd)
    if err != nil {
      return err
    }
  }

  return
}

// Kill instance will kill a single instance in the environment
func (pool AmazonSpawnPool) Shrink() (err error) {
  // This should uh, kill one instance at random
  // Doesn't matter if it's the leader, but it might be nice to avoid that

  return errors.New("Not implemented")
}

// This will set-up a new spawn pool, if it doesn't exist
// For AWS, we'll track this by security groups
func (pool *AmazonSpawnPool) findOrCreateSecurityGroup() (ec2.SecurityGroup, error) {
  sgName := fmt.Sprintf("fiddler-%s", pool.name)
  groups := []ec2.SecurityGroup{ec2.SecurityGroup{Name: sgName}}
  
  e := ec2.New(pool.auth, aws.USEast)

  sgResp, sgErr := e.SecurityGroups(groups, nil)
  if sgErr == nil { // We found the security group
    return sgResp.Groups[0].SecurityGroup, nil
  } else {
    log.Println("Creating security group:", pool.name)
    resp, err := e.CreateSecurityGroup(sgName, fmt.Sprintf("Fiddler env for %s", sgName))
    if err != nil {
      return ec2.SecurityGroup{}, err
    }

    // For now, we're going to open up 4001/7001 to the group
    // and ports 21, 80 to the world
    // This should have a config option
    sg := resp.SecurityGroup

    openPorts := []int{22, 80}
    perms := make([]ec2.IPPerm, 2+len(openPorts))
    perms[0] = ec2.IPPerm{Protocol: "tcp", FromPort: 4001, ToPort: 4001, SourceGroups: []ec2.UserSecurityGroup{ec2.UserSecurityGroup{Id: sg.Id}}}
    perms[1] = ec2.IPPerm{Protocol: "tcp", FromPort: 7001, ToPort: 7001, SourceGroups: []ec2.UserSecurityGroup{ec2.UserSecurityGroup{Id: sg.Id}}}
    for i, port := range openPorts {
      perms[i+2] = ec2.IPPerm{Protocol: "tcp", FromPort: port, ToPort: port, SourceIPs: []string{"0.0.0.0/0"}}
    }

    _, err = e.AuthorizeSecurityGroup(sg, perms)
    if err != nil {
      // Delete unauthorized SG?
      return ec2.SecurityGroup{}, err
    }

    return sg, err
  }