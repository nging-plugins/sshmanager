package handler

import (
	"github.com/coscms/webcore/library/navigate"
	"github.com/webx-top/echo"
)

var LeftNavigate = &navigate.Item{
	Display: true,
	Name:    echo.T(`SSH管理`),
	Action:  `term`,
	Icon:    `terminal`,
	Children: &navigate.List{
		{
			Display: true,
			Name:    echo.T(`账号管理`),
			Action:  `account`,
		},
		{
			Display: true,
			Name:    echo.T(`添加账号`),
			Action:  `account_add`,
			Icon:    `plus`,
		},
		{
			Display: true,
			Name:    echo.T(`分组管理`),
			Action:  `group`,
		},
		{
			Display: true,
			Name:    echo.T(`添加分组`),
			Action:  `group_add`,
			Icon:    `plus`,
		},
		{
			Display: false,
			Name:    echo.T(`修改账号`),
			Action:  `account_edit`,
			Icon:    ``,
		},
		{
			Display: false,
			Name:    echo.T(`删除账号`),
			Action:  `account_delete`,
			Icon:    ``,
		},
		{
			Display: false,
			Name:    echo.T(`修改分组`),
			Action:  `group_edit`,
			Icon:    ``,
		},
		{
			Display: false,
			Name:    echo.T(`删除分组`),
			Action:  `group_delete`,
			Icon:    ``,
		},
		{
			Display: false,
			Name:    echo.T(`SSH操作`),
			Action:  `client`,
			Icon:    ``,
		},
		{
			Display: false,
			Name:    echo.T(`SFTP操作`),
			Action:  `sftp`,
			Icon:    ``,
		},
	},
}
