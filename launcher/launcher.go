package launcher

import (
  "github.com/hayesgm/fiddler/config"
  "github.com/hayesgm/fiddler/piper"
  "log"
  "os/exec"
)

// TODO: which docker
// This will launch a docker for the configuration
// We're going to set-up links on the new instance as unix sockets
func Launch(run *config.RunConf, links []piper.Pipe) (cmd *exec.Cmd, err error) {
  log.Printf("Launching %#v", run)
  
  runArgsLen := 0

  if run.Args != nil {
    runArgsLen = len(run.Args)
  }

  args := make([]string, 4+runArgsLen)
  args[0] = "/usr/bin/docker"
  args[1] = "run"
  args[2] = run.Container

  j := 3

  if len(run.Exec) > 0 {
    args[j] = run.Exec
    j++
  }

  for i := 0; i < runArgsLen; i++ {
    args[i+j] = run.Args[i]
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