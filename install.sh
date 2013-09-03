#! /bin/sh

echo "We will need sudo to write systemd service"
sudo echo "Continuing installation"

cd ~
wget https://github.com/hayesgm/fiddler/raw/master/bin/fiddler.linux.tar.gz && tar -xvf fiddler.linux.tar.gz && sudo ./fiddler.linux -i -c $1