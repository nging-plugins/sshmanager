package sshmanager

import (
	"github.com/coscms/webcore/library/module"
	"github.com/nging-plugins/sshmanager/application/handler"
)

var LeftNavigate = handler.LeftNavigate

func RegisterNavigate(nc module.Navigate) {
	nc.Backend().AddLeftItems(-1, LeftNavigate)
}
