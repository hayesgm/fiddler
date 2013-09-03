package launcher

import (
  "github.com/hayesgm/fiddler/config"
  "log"
  "os/exec"
)

// TODO: which docker
func Launch(docker *config.DockerConf) (cmd *exec.Cmd, err error) {
  log.Printf("Launching %#v", docker)
  
  args := make([]string, 2+len(docker.Args))
  args[0] = "run"
  args[1] = docker.Container
  for i := 0; i < len(docker.Args); i++ {
    args[i+2] = docker.Args[i]
  }
  cmd = &exec.Cmd{Path: "/usr/bin/docker", Args: args}
  
  err = cmd.Start()
  if err != nil {
    return
  }
  
  log.Printf("Successfully launched container...")

  return
}