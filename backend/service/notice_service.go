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

	"gorm.io/gorm"

	"gitee.com/ivfzhou/csms/backend/consts"
	"gitee.com/ivfzhou/csms/backend/protocol"
	"gitee.com/ivfzhou/csms/comm/conn"
	"gitee.com/ivfzhou/csms/comm/errs"
	"gitee.com/ivfzhou/csms/comm/log"
	"gitee.com/ivfzhou/csms/comm/model"
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
