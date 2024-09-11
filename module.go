package sshmanager

import (
	"github.com/coscms/webcore/library/module"

	"github.com/nging-plugins/sshmanager/application/handler"
	"github.com/nging-plugins/sshmanager/application/library/setup"
)

const ID = `term`

var Module = module.Module{
	TemplatePath: map[string]string{
		ID: `sshmanager/template/backend`,
	},
	AssetsPath:    []string{},
	SQLCollection: setup.RegisterSQL,
	Navigate:      RegisterNavigate,
	Route:         handler.RegisterRoute,
	DBSchemaVer:   0.1000,
}
