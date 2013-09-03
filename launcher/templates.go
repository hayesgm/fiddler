package launcher

var FiddlerTemplate = `
[Unit]
Description=Fiddler
After=docker.service

[Service]
Restart=always
ExecStart={{.Exec}} -c {{.Conf}}

[Install]
WantedBy=local.target
`

var DockerTemplate = `
[Unit]
Description=Container
After=docker.service

[Service]
Restart=always
ExecStart=/usr/bin/docker run -d {{.Container}} {{.Run}} {{range .Args}} {{.}} {{end}}

[Install]
WantedBy=local.target
`