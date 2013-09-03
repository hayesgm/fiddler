package launcher

import (
  "os"
  "io"
  "bitbucket.org/kardianos/osext"
  "path"
  "log"
  "text/template"
  "bytes"
  "os/exec"
)

type Service struct {
  name string
  contents []byte
}

type ServiceSettings struct {
  Exec string
  Conf string
}

// Installs Fiddler to run on systemd
// 1) This is going to copy our executable to a more suitable location (/usr/bin/fiddler)
// 2) Then, we'll install a service for ourselves
// 3) And we'll kick off that service
func InstallFiddler(conf string) (err error) {
  // err = copyToBin()
  // if err != nil {
  //  return
  // }
  src, err := osext.Executable()
  if err != nil {
    return
  }

  tmpl, err := template.New("fiddler.service").Parse(FiddlerTemplate)
  if err != nil {
    return
  }
  var contents bytes.Buffer

  settings := ServiceSettings{Exec: src, Conf: conf}
  tmpl.Execute(&contents, settings)

  service := Service{name: "fiddler.service", contents: contents.Bytes()}
  err = installService(service)
  if err != nil {
    return
  }

  err = restartServices()
  if err != nil {
    return
  }

  return
}

func copyToBin() (err error) {
  src, err := osext.Executable()
  if err != nil {
    return
  }

  dst := "/usr/bin/fiddler"

  log.Println("Copying", src, "to", dst, "...")

  sf, err := os.OpenFile(src, os.O_RDONLY, 0666)
  if err != nil {
    return
  }
  defer sf.Close()

  df, err := os.OpenFile(dst, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0744)
  if err != nil {
    return
  }
  defer df.Close()

  _, err = io.Copy(df, sf)
  if err != nil {
    return 
  }

  log.Println("Done.")

  return
}

func installService(service Service) (err error) {
  // We're going to write this service
  log.Printf("Installing Fiddler Service, %#v\n", string(service.contents))

  sf, err := os.OpenFile(path.Join("/media/state/units",service.name), os.O_TRUNC|os.O_CREATE, 0644)
  if err != nil {
    return err
  }
  defer sf.Close()

  _, err = sf.Write(service.contents)
  if err != nil {
    return
  }

  log.Println("Done.")
  return
}

func restartServices() (err error) {
  log.Println("Restarting Serviced Services...")
  cmd := exec.Command("systemctl", "restart", "local-enable.service")
  err = cmd.Run()
  if err != nil {
    return
  }

  log.Println("Done.")
  return
}