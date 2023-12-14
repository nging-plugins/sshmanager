package cloudbackup

import (
	"context"
	"io"
	"os"
	"path"
	"path/filepath"

	nd "github.com/admpub/nging/v5/application/dbschema"
	"github.com/admpub/nging/v5/application/library/cloudbackup"
	"github.com/admpub/nging/v5/application/library/sftpmanager"
	"github.com/admpub/nging/v5/application/model"
	"github.com/nging-plugins/sshmanager/application/dbschema"
	sshconf "github.com/nging-plugins/sshmanager/application/library/config"
	"github.com/webx-top/com"
	"github.com/webx-top/echo"
)

func init() {
	cloudbackup.Register(model.StorageEngineSFTP, newStorageSFTP, sftpForms, `SFTP`)
}

func newStorageSFTP(ctx echo.Context, cfg nd.NgingCloudBackup) (cloudbackup.Storager, error) {
	m := dbschema.NewNgingSshUser(ctx)
	err := m.Get(nil, `id`, cfg.DestStorage)
	if err != nil {
		return nil, err
	}
	conf := sshconf.ToSFTPConfig(m)
	return NewStorageSFTP(conf), nil
}

var sftpForms = []cloudbackup.Form{
	{Type: `text`, Label: `SSH账号`, Name: `destStorage`, Required: true},
}

func NewStorageSFTP(cfg sftpmanager.Config) cloudbackup.Storager {
	return &StorageSFTP{cfg: cfg}
}

type StorageSFTP struct {
	cfg  sftpmanager.Config
	conn *sftpmanager.SftpManager
}

func (s *StorageSFTP) Connect() (err error) {
	s.conn = sftpmanager.New(sftpmanager.DefaultConnector, &s.cfg, 0)
	s.conn.Client()
	if s.conn.ConnError() != nil {
		err = s.conn.ConnError()
	}
	return
}

func (s *StorageSFTP) Put(ctx context.Context, reader io.Reader, ppath string, size int64) (err error) {
	s.conn.MkdirAll(ctx, path.Dir(ppath))
	err = s.conn.Put(ctx, reader, ppath, size)
	return
}

func (s *StorageSFTP) Download(ctx context.Context, ppath string, w io.Writer) error {
	c := s.conn.Client()
	resp, err := c.Open(ppath)
	if err != nil {
		return err
	}
	defer resp.Close()
	_, err = io.Copy(w, resp)
	return err
}

func (s *StorageSFTP) Restore(ctx context.Context, ppath string, destpath string) error {
	c := s.conn.Client()
	resp, err := c.Open(ppath)
	if err != nil {
		return err
	}
	defer resp.Close()
	stat, err := resp.Stat()
	if err != nil {
		return err
	}
	if !stat.IsDir() {
		return cloudbackup.DownloadFile(s, ctx, ppath, destpath)
	}
	dirs, err := c.ReadDir(ppath)
	if err != nil {
		return err
	}
	for _, dir := range dirs {
		spath := path.Join(ppath, dir.Name())
		dest := filepath.Join(destpath, dir.Name())
		if dir.IsDir() {
			err = com.MkdirAll(dest, os.ModePerm)
			if err == nil {
				err = s.Restore(ctx, spath, dest)
			}
		} else {
			err = cloudbackup.DownloadFile(s, ctx, spath, dest)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *StorageSFTP) RemoveDir(ctx context.Context, ppath string) error {
	return s.conn.RemoveDir(ppath)
}

func (s *StorageSFTP) Remove(ctx context.Context, ppath string) error {
	return s.conn.Remove(ppath)
}

func (s *StorageSFTP) Close() (err error) {
	if s.conn == nil {
		return
	}
	err = s.conn.Close()
	return
}
