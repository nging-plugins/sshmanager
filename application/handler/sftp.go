/*
   Nging is a toolbox for webmasters
   Copyright (C) 2018-present  Wenhui Shen <swh@admpub.com>

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published
   by the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package handler

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	uploadClient "github.com/webx-top/client/upload"
	uploadDropzone "github.com/webx-top/client/upload/driver/dropzone"
	"github.com/webx-top/com"
	"github.com/webx-top/echo"

	"github.com/coscms/webcore/library/backend"
	"github.com/coscms/webcore/library/common"
	"github.com/coscms/webcore/library/config"
	"github.com/coscms/webcore/library/filemanager"
	"github.com/coscms/webcore/library/notice"
	"github.com/coscms/webcore/library/respond"
	"github.com/coscms/webcore/library/sftpmanager"
	uploadChunk "github.com/coscms/webcore/registry/upload/chunk"

	"github.com/nging-plugins/sshmanager/application/dbschema"
	sshconf "github.com/nging-plugins/sshmanager/application/library/config"
	"github.com/nging-plugins/sshmanager/application/model"
)

func sftpConfig(m *dbschema.NgingSshUser) sftpmanager.Config {
	return sshconf.ToSFTPConfig(m)
}

func SftpSearch(ctx echo.Context, id uint) error {
	m := model.NewSshUser(ctx)
	err := m.Get(nil, `id`, id)
	if err != nil {
		return err
	}
	mgr, err := getCachedSFTPClient(m.NgingSshUser)
	if err != nil {
		return err
	}
	query := ctx.Form(`query`)
	num := ctx.Formx(`size`, `10`).Int()
	if num <= 0 {
		num = 10
	}
	client := mgr.Client()
	if mgr.ConnError() != nil {
		deleteCachedSFTPClient(m.NgingSshUser.Id)
		return mgr.ConnError()
	}
	paths := sftpmanager.Search(client, query, ctx.Form(`type`), num)
	data := ctx.Data().SetData(paths)
	return ctx.JSON(data)
}

func Sftp(ctx echo.Context) error {
	ctx.Set(`activeURL`, `/term/account`)
	id := ctx.Formx(`id`).Uint()
	m := model.NewSshUser(ctx)
	err := m.Get(nil, `id`, id)
	if err != nil {
		return err
	}
	var mgr *sftpmanager.SftpManager
	mgr, err = getCachedSFTPClient(m.NgingSshUser)
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			return
		}
		if strings.Contains(err.Error(), `connection lost`) {
			deleteCachedSFTPClient(m.NgingSshUser.Id)
		}
	}()
	ppath := ctx.Form(`path`)
	do := ctx.Form(`do`)
	parentPath := ppath
	if len(ppath) == 0 {
		if len(m.SftpRootDir) > 0 {
			ppath = m.SftpRootDir
		} else {
			ppath = `/`
		}
	} else {
		parentPath = path.Dir(ppath)
	}
	user := backend.User(ctx)
	switch do {
	case `edit`:
		data := ctx.Data()
		if _, ok := config.FromFile().Sys.Editable(ppath); !ok {
			data.SetInfo(ctx.T(`此文件不能在线编辑`), 0)
		} else {
			content := ctx.Form(`content`)
			encoding := ctx.Form(`encoding`)
			var dat interface{}
			dat, err = mgr.Edit(ctx, ppath, content, encoding)
			if err != nil {
				data.SetInfo(err.Error(), 0)
			} else {
				if ctx.IsPost() {
					data.SetInfo(ctx.T(`保存成功`), 1)
				}
				data.SetData(dat, 1)
			}
		}
		return ctx.JSON(data)
	case `mkdir`:
		data := ctx.Data()
		newName := ctx.Form(`name`)
		if len(newName) == 0 {
			data.SetInfo(ctx.T(`请输入文件夹名`), 0)
		} else {
			err = mgr.Mkdir(ctx, ppath, newName)
			if err != nil {
				data.SetError(err)
			}
			if data.GetCode() == 1 {
				data.SetInfo(ctx.T(`创建成功`))
			}
		}
		return ctx.JSON(data)
	case `rename`:
		data := ctx.Data()
		newName := ctx.Form(`name`)
		err = mgr.Rename(ctx, ppath, newName)
		if err != nil {
			data.SetInfo(err.Error(), 0)
		} else {
			data.SetInfo(ctx.T(`重命名成功`), 1)
		}
		return ctx.JSON(data)
	case `chown`:
		data := ctx.Data()
		uid := ctx.Formx(`uid`).Int()
		gid := ctx.Formx(`gid`).Int()
		err = mgr.Chown(ctx, ppath, uid, gid)
		if err != nil {
			data.SetInfo(err.Error(), 0)
		} else {
			data.SetInfo(ctx.T(`操作成功`), 1)
		}
		return ctx.JSON(data)
	case `chmod`:
		data := ctx.Data()
		mode := ctx.Formx(`mode`).Uint32() //0777 etc...
		err = mgr.Chmod(ctx, ppath, os.FileMode(mode))
		if err != nil {
			data.SetInfo(err.Error(), 0)
		} else {
			data.SetInfo(ctx.T(`操作成功`), 1)
		}
		return ctx.JSON(data)
	case `search`:
		prefix := ctx.Form(`query`)
		num := ctx.Formx(`size`, `10`).Int()
		if num <= 0 {
			num = 10
		}
		paths := mgr.Search(ppath, prefix, num)
		data := ctx.Data().SetData(paths)
		return ctx.JSON(data)
	case `delete`:
		paths := ctx.FormValues(`path`)
		for _, _path := range paths {
			if len(_path) == 0 {
				continue
			}
			_path = path.Clean(_path)
			err = mgr.Remove(_path)
			if err != nil {
				break
			}
		}
		if err != nil {
			common.SendFail(ctx, err.Error())
		}
		next := ctx.Query(`next`)
		if len(next) == 0 {
			next = ctx.Referer()
			if len(next) == 0 {
				next = ctx.Request().URL().Path() + fmt.Sprintf(`?id=%d&path=%s`, id, com.URLEncode(path.Dir(ppath)))
			}
		}
		return ctx.Redirect(next)
	case `upload`:
		var cu *uploadClient.ChunkUpload
		var opts []uploadClient.ChunkInfoOpter
		if user != nil {
			cu = uploadChunk.NewUploader(fmt.Sprintf(`user/%d`, user.Id))
			opts = append(opts, uploadClient.OptChunkInfoMapping(uploadDropzone.MappingChunkInfo))
			np := notice.NewP(ctx, `uploadToSFTP`, user.Username, context.Background())
			ctx.Internal().Set(`noticer`, np)
		}
		err = mgr.Upload(ctx, ppath, cu, opts...)
		if err != nil {
			user := backend.User(ctx)
			if user != nil {
				notice.OpenMessage(user.Username, `upload`)
				notice.Send(user.Username, notice.NewMessageWithValue(`upload`, ctx.T(`文件上传出错`), err.Error()))
			}
		}
		return respond.Dropzone(ctx, err, nil)
	default:
		var dirs []os.FileInfo
		var exit bool
		err, exit, dirs = mgr.List(ctx, ppath)
		if exit {
			return err
		}
		ctx.Set(`dirs`, dirs)
	}
	ctx.Set(`parentPath`, parentPath)
	ctx.Set(`path`, ppath)
	pathPrefix := ppath
	if ppath != `/` {
		pathPrefix = ppath + `/`
	}
	pathSlice := strings.Split(strings.Trim(pathPrefix, `/`), `/`)
	pathLinks := make(echo.KVList, len(pathSlice))
	encodedSep := filemanager.EncodedSep
	urlPrefix := ctx.Request().URL().Path() + fmt.Sprintf(`?id=%d&path=`, id) + encodedSep
	for k, v := range pathSlice {
		urlPrefix += com.URLEncode(v)
		pathLinks[k] = &echo.KV{K: v, V: urlPrefix}
		urlPrefix += encodedSep
	}
	ctx.Set(`pathLinks`, pathLinks)
	ctx.Set(`pathPrefix`, pathPrefix)
	ctx.SetFunc(`Editable`, func(fileName string) bool {
		_, ok := config.FromFile().Sys.Editable(fileName)
		return ok
	})
	ctx.SetFunc(`Playable`, func(fileName string) string {
		mime, _ := config.FromFile().Sys.Playable(fileName)
		return mime
	})
	ctx.Set(`data`, m.NgingSshUser)
	return ctx.Render(`term/sftp`, err)
}

func AutoCompletePath(ctx echo.Context) (bool, error) {
	sshAccountID := ctx.Formx(`sshAccountId`).Uint()
	if sshAccountID > 0 {
		check, _ := ctx.Funcs()[`CheckPerm`].(func(string) error)
		data := ctx.Data()
		if check == nil {
			data.SetData([]string{})
			return true, ctx.JSON(data)
		}
		if err := check(`manager/command_add`); err != nil {
			return true, err
		}
		if err := check(`manager/command_edit`); err != nil {
			return true, err
		}
		return true, SftpSearch(ctx, sshAccountID)
	}
	return false, nil
}
