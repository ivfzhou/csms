# 获取等待初始化测试机的任务。
select t2.name                                                             `应用名`,
       t2.app_id                                                           `应用 ID`,
       t1.job_id                                                           `任务 ID`,
       case t1.source when 1 then t3.name_en when 2 then t4.account_id end `任务提交人`,
       t1.file_id                                                          `文件 ID`,
       timestampdiff(minute, ifnull(t1.updated_time, t1.created_time),
                     convert_tz(current_timestamp(), '+00:00', '+08:00'))  `等待时间（分钟）`,
       t1.test_system                                                      `测试系统`,
       t1.log                                                              `日志`
from t_whql_job t1
         left join t_app t2 on t2.id = t1.app_id
         left join t_user t3 on t3.id = t1.user_id
         left join t_api_account t4 on t4.id = t1.user_id
where t1.status = 1
order by `等待时间（分钟）` desc, t1.id desc
limit 10;

# 获取正在初始化测试机的任务。
select t2.name                                                                                     `应用名`,
       t2.app_id                                                                                   `应用 ID`,
       t1.job_id                                                                                   `任务 ID`,
       case t1.source when 1 then t3.name_en when 2 then t4.account_id end                         `任务提交人`,
       t1.file_id                                                                                  `文件 ID`,
       timestampdiff(minute, t1.updated_time, convert_tz(current_timestamp(), '+00:00', '+08:00')) `等待时间（分钟）`,
       t1.test_system                                                                              `测试系统`,
       t1.log                                                                                      `日志`
from t_whql_job t1
         left join t_app t2 on t2.id = t1.app_id
         left join t_user t3 on t3.id = t1.user_id
         left join t_api_account t4 on t4.id = t1.user_id
where t1.status = 2
order by `等待时间（分钟）` desc, t1.id desc
limit 10;

# 获取等待控制器调度测试的任务。
select t2.name                                                                                     `应用名`,
       t2.app_id                                                                                   `应用 ID`,
       t1.job_id                                                                                   `任务 ID`,
       case t1.source when 1 then t3.name_en when 2 then t4.account_id end                         `任务提交人`,
       t1.file_id                                                                                  `文件 ID`,
       timestampdiff(minute, t1.updated_time, convert_tz(current_timestamp(), '+00:00', '+08:00')) `等待时间（分钟）`,
       t1.test_system                                                                              `测试系统`,
       t1.log                                                                                      `日志`
from t_whql_job t1
         left join t_app t2 on t2.id = t1.app_id
         left join t_user t3 on t3.id = t1.user_id
         left join t_api_account t4 on t4.id = t1.user_id
where t1.status = 3
order by `等待时间（分钟）` desc, t1.id desc
limit 10;

# 获取正在调度测试的任务。
select t2.name                                                                                     `应用名`,
       t2.app_id                                                                                   `应用 ID`,
       t1.job_id                                                                                   `任务 ID`,
       case t1.source when 1 then t3.name_en when 2 then t4.account_id end                         `任务提交人`,
       t1.file_id                                                                                  `文件 ID`,
       timestampdiff(minute, t1.updated_time, convert_tz(current_timestamp(), '+00:00', '+08:00')) `等待时间（分钟）`,
       t1.test_system                                                                              `测试系统`,
       t1.log                                                                                      `日志`
from t_whql_job t1
         left join t_app t2 on t2.id = t1.app_id
         left join t_user t3 on t3.id = t1.user_id
         left join t_api_account t4 on t4.id = t1.user_id
where t1.status = 4
order by `等待时间（分钟）` desc, t1.id desc
limit 10;

# 获取正在测试的任务。
select t2.name                                                                                     `应用名`,
       t2.app_id                                                                                   `应用 ID`,
       t1.job_id                                                                                   `任务 ID`,
       case t1.source when 1 then t3.name_en when 2 then t4.account_id end                         `任务提交人`,
       t1.file_id                                                                                  `文件 ID`,
       timestampdiff(minute, t1.updated_time, convert_tz(current_timestamp(), '+00:00', '+08:00')) `等待时间（分钟）`,
       t1.test_system                                                                              `测试系统`,
       t1.log                                                                                      `日志`
from t_whql_job t1
         left join t_app t2 on t2.id = t1.app_id
         left join t_user t3 on t3.id = t1.user_id
         left join t_api_account t4 on t4.id = t1.user_id
where t1.status = 5
order by `等待时间（分钟）` desc, t1.id desc
limit 10;