package launcher

var FiddlerTemplate = `
[Unit]
Description=Fiddler
After=docker.service

[Service]
Restart=always
ExecStart={{.Exec}} -l -c {{.Conf}}

[Install]
WantedBy=local.target
`

var DockerTemplate = `
[Unit]
Description=Docker
After=docker.service

[Service]
Restart=always
ExecStart=/usr/bin/docker run {{.Container}} {{.Run}} {{range .Args}} {{.}} {{end}}

[Install]
WantedBy=local.target
`