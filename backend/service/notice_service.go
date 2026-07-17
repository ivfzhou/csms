/*
 * Copyright (c) 2024 ivfzhou
 * csms is licensed under Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *          http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
 * EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
 * MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
 * See the Mulan PSL v2 for more details.
 */

package service

import (
	"context"
	"errors"
	"time"

	"gorm.io/gen"
	"gorm.io/gorm"

	"gitee.com/ivfzhou/csms/backend/consts"
	"gitee.com/ivfzhou/csms/backend/protocol"
	"gitee.com/ivfzhou/csms/comm/conn"
	"gitee.com/ivfzhou/csms/comm/ctxs"
	"gitee.com/ivfzhou/csms/comm/errs"
	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/model"
	"gitee.com/ivfzhou/csms/comm/util"
)

// NoticeWebLast 获取通知。
func NoticeWebLast(ctx context.Context) (rsp *protocol.NoticeWebLastRsp, err error) {
	// 查库，获取最新的通知。
	var notice *model.Notice
	{
		noticeDo := conn.MySQLClient(ctx).Notice
		notice, err = noticeDo.WithContext(ctx).Where(
			noticeDo.ActivatedTime.Lt(time.Now()),
			noticeDo.ExpiredTime.Gt(time.Now()),
		).Order(
			noticeDo.CreatedTime.Desc(),
			noticeDo.ID.Desc(),
		).Limit(1).Take()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				rsp = &protocol.NoticeWebLastRsp{}
				err = nil
				return
			}
			log.Error(ctx, "failed to retrieve notice from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	rsp = &protocol.NoticeWebLastRsp{Message: notice.Content}

	return
}

// NoticeWebAdd 添加通知。
func NoticeWebAdd(ctx context.Context, req *protocol.NoticeWebAddReq) (err error) {
	// 获取上下文信息。
	var user *model.User
	{
		log.Info(ctx, "get context information")
		user = ctxs.User(ctx)
		if user == nil {
			log.Warn(ctx, "unknown context", user)
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 将通知保存进数据库。
	{
		log.Info(ctx, "add notice to database")
		noticeDo := conn.MySQLClient(ctx).Notice
		err = noticeDo.WithContext(ctx).Create(&model.Notice{
			Content:       req.Message,
			UserID:        user.ID,
			ExpiredTime:   time.Time(req.ExpiredTime),
			ActivatedTime: time.Time(req.ActivatedTime),
			CreatedTime:   time.Now(),
		})
		if err != nil {
			log.Error(ctx, "failed to add notice to database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	return
}

// NoticeWebList 通知列表。
func NoticeWebList(ctx context.Context) (rsp *protocol.NoticeWebListRsp, err error) {
	// 获取上下文信息。
	var user *model.User
	{
		log.Info(ctx, "get context information")
		user = ctxs.User(ctx)
		if user == nil {
			log.Warn(ctx, "unknown context", user)
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 查库获取通知。
	var notices []*model.Notice
	{
		log.Info(ctx, "get notices from database")
		noticeDo := conn.MySQLClient(ctx).Notice
		notices, err = noticeDo.WithContext(ctx).Order(noticeDo.ID.Desc()).Find()
		if err != nil {
			log.Error(ctx, "failed to retrieve notices from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
	}

	// 获取用户名。
	var userIDToName map[int]string
	{
		userIDs := util.ListToUnique(notices, func(v *model.Notice) int { return v.UserID })
		if len(userIDs) > 0 {
			log.Info(ctx, "get user names")
			userIDToName, err = GetUserNamesByIDs(ctx, userIDs)
			if err != nil {
				return
			}
		} else {
			userIDToName = make(map[int]string)
		}
	}

	// 组装数据。
	{
		list := make([]*protocol.NoticeWebListItem, len(notices))
		for i, v := range notices {
			list[i] = &protocol.NoticeWebListItem{
				ID:            v.ID,
				Message:       v.Content,
				User:          userIDToName[v.UserID],
				ActivatedTime: formatTime(&v.ActivatedTime),
				ExpiredTime:   formatTime(&v.ExpiredTime),
				CreatedTime:   formatTime(&v.CreatedTime),
			}
		}
		rsp = &protocol.NoticeWebListRsp{List: list}
	}

	return
}

// NoticeWebRemove 删除通知。
func NoticeWebRemove(ctx context.Context, req *protocol.NoticeWebRemoveReq) (err error) {
	// 获取上下文信息。
	var user *model.User
	{
		log.Info(ctx, "get context information")
		user = ctxs.User(ctx)
		if user == nil {
			log.Warn(ctx, "unknown context", user)
			err = errs.New(consts.ErrSystem)
			return
		}
	}

	// 删除数据库中的公告内容。
	{
		log.Info(ctx, "remove notice from database")
		noticeDo := conn.MySQLTxClient(ctx).Notice
		var sqlResult gen.ResultInfo
		sqlResult, err = noticeDo.WithContext(ctx).Where(noticeDo.ID.Eq(req.ID)).Delete()
		if err != nil {
			log.Error(ctx, "failed to remove notice from database", err)
			err = errs.NewWithError(consts.ErrSystem, err)
			return
		}
		if sqlResult.RowsAffected <= 0 {
			log.Warn(ctx, "failed to remove notice from database", sqlResult.RowsAffected, req.ID)
		}
	}

	return
}
