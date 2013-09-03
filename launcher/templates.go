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