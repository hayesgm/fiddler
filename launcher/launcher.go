package launcher

import (
  "github.com/hayesgm/fiddler/config"
  "log"
  "os/exec"
)

// TODO: which docker
func Launch(docker *config.DockerConf) (cmd *exec.Cmd, err error) {
  log.Printf("Launching %#v", docker)
  
  args := make([]string, 4+len(docker.Args))
  args[0] = "/usr/bin/docker"
  args[1] = "run"
  args[2] = docker.Container

  j := 3

  if len(docker.Run) > 0 {
    args[j] = docker.Run
    j++
  }

  for i := 0; i < len(docker.Args); i++ {
    args[i+j] = docker.Args[i]
  }

  cmd = &exec.Cmd{Path: "/bin/echo", Args: args}
  // For now, we'll show this output specifically
  // We may want to pipe this to a file
  //cmd.Stdout = os.Stdout
  //cmd.Stderr = os.Stderr
  //log.Printf("Running cmd: %#v\n", cmd)
  err = cmd.Start()
  if err != nil {
    return
  }
  
  log.Printf("Successfully launched container...")
  return
}