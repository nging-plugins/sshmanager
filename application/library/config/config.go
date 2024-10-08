package config

import (
	"github.com/coscms/webcore/library/sftpmanager"
	"github.com/nging-plugins/sshmanager/application/dbschema"
)

func ToSFTPConfig(m *dbschema.NgingSshUser) sftpmanager.Config {
	return sftpmanager.Config{
		Host:       m.Host,
		Port:       m.Port,
		Username:   m.Username,
		Password:   m.Password,
		PrivateKey: m.PrivateKey,
		Passphrase: m.Passphrase,
		Charset:    m.Charset,
	}
}
