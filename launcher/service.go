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
  "github.com/hayesgm/fiddler/config"
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
func InstallFiddler(c string, conf config.FiddlerConf, launch bool) (err error) {
  // err = copyToBin()
  // if err != nil {
  //  return
  // }
  src, err := osext.Executable()
  if err != nil {
    return
  }

  settings := ServiceSettings{Exec: src, Conf: c}

  err = installServiceFromTemplate("fiddler.service", FiddlerTemplate, &settings)
  if err != nil {
    return
  }

  if !launch {
    err = installServiceFromTemplate("docker.service", DockerTemplate, conf.Docker)
    if err != nil {
      return
    }
  }

  err = restartServices()
  if err != nil {
    return
  }

  return
}

func installServiceFromTemplate(name string, templateText string, settings interface{}) (err error) {
  tmpl, err := template.New(name).Parse(templateText)
  if err != nil {
    return
  }
  var contents bytes.Buffer

  tmpl.Execute(&contents, settings)

  service := Service{name: name, contents: contents.Bytes()}
  err = installService(service)
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

  sf, err := os.OpenFile(path.Join("/media/state/units",service.name), os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
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