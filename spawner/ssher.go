package spawner

import (
  "code.google.com/p/go.crypto/ssh"
  "crypto"
  "crypto/rsa"
  "crypto/x509"
  "encoding/pem"
  "io"
)

type keychain struct {
  key *rsa.PrivateKey
}

func (k *keychain) Key(i int) (interface{}, error) {
  if i != 0 {
    return nil, nil
  }
  return &k.key.PublicKey, nil
}

func (k *keychain) Sign(i int, rand io.Reader, data []byte) (sig []byte, err error) {
  hashFunc := crypto.SHA1
  h := hashFunc.New()
  h.Write(data)
  digest := h.Sum(nil)
  return rsa.SignPKCS1v15(rand, k.key, hashFunc, digest)
}

func sshCommand(ip, user, privateKey, command string) (err error) {
  block, _ := pem.Decode([]byte(privateKey))
  rsakey, _ := x509.ParsePKCS1PrivateKey(block.Bytes)
  clientKey := &keychain{rsakey}
  clientConfig := &ssh.ClientConfig {
    User: user,
    Auth: []ssh.ClientAuth{
      ssh.ClientAuthKeyring(clientKey),
    },
  }
  client, err := ssh.Dial("tcp",ip, clientConfig)
  if err != nil {
    return err
  }
  session, err := client.NewSession()
  if err != nil {
    return err
  }
  defer session.Close()
  if err := session.Run(command); err != nil {
    return err
  }

  return
}