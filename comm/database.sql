/*
 * Copyright (c) 2025 ivfzhou
 * csms is licensed under Mulan PSL v2.
 * You can use this software according to the terms and conditions of the Mulan PSL v2.
 * You may obtain a copy of Mulan PSL v2 at:
 *          http://license.coscl.org.cn/MulanPSL2
 * THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
 * EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
 * MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
 * See the Mulan PSL v2 for more details.
 */

create database if not exists db_csms character set utf8mb4 collate utf8mb4_unicode_ci;

use db_csms;

create table if not exists t_user
(
    id              int unsigned primary key auto_increment comment '自增主码',
    name_en         varchar(32)   not null comment '英文名',
    name_zh         varchar(16)   not null comment '中文名',
    avatar_file_id  char(38)      not null comment '头像文件 ID',
    department      varchar(1024) not null comment '部门组织，以/分隔',
    password_digest char(32)      not null comment '密码 MD5',
    password_salt   char(128)     not null comment '密码盐值',
    created_time    timestamp     not null comment '注册时间',
    updated_time    timestamp comment '更新时间',
    unique index uidx_name_en (name_en) using btree
) engine innodb comment '用户信息表';

create table if not exists t_app
(
    id           int unsigned primary key auto_increment comment '自增主码',
    app_id       char(32)         not null comment '唯一标识',
    user_id      int unsigned     not null comment '创建人 ID',
    name         varchar(64)      not null comment '应用名',
    logo_file_id char(38)         not null comment '图标文件 ID',
    platform     tinyint unsigned not null comment '平台，1=Windows、2=Apple、3=Android',
    status       tinyint unsigned not null comment '状态，1=有效、2=无效、3=审批中、4=审批驳回',
    created_time timestamp        not null comment '注册时间',
    updated_time timestamp comment '更新时间',
    unique index uidx_app_id (app_id) using btree,
    index idx_status (status) using btree
) engine innodb comment '应用信息表';

create table if not exists t_user_role
(
    id      int unsigned primary key auto_increment comment '自增主码',
    app_id  int unsigned comment '所属应用 ID',
    user_id int unsigned     not null comment '用户 ID',
    role    tinyint unsigned not null comment '角色，1=系统运营员、2=应用管理员、3=应用成员、4=可使用签名服务',
    unique index uidx_app_id_user_id_role (user_id, role, app_id) using btree,
    index idx_user_id (user_id) using btree
) engine innodb comment '用户权限表';

create table if not exists t_event
(
    id           int unsigned primary key auto_increment comment '自增主码',
    app_id       int unsigned     not null comment '所属应用 ID',
    user_id      int unsigned     not null comment '触发人 ID',
    type         tinyint unsigned not null comment '类型，见表 3-3',
    created_time timestamp        not null comment '事件发生时间',
    source       tinyint unsigned not null comment '来源，1=页面、2=OpenAPI',
    content      text comment '事件内容',
    index idx_created_time_type_app_id (created_time, type, app_id) using btree
) engine innodb comment '应用事件表';

create table if not exists t_todo
(
    id              int unsigned primary key auto_increment comment '自增主码',
    app_id          int unsigned     not null comment '所属应用 ID',
    type            tinyint unsigned not null comment '类型，见表 3-4',
    applier_id      int unsigned     not null comment '申请人 ID',
    approver_id     int unsigned comment '审批人 ID',
    candidates      varchar(256)     not null comment '待办可审批的人 ID，英文逗号分隔',
    apply_reason    varchar(256) comment '申请理由',
    approve_message varchar(256) comment '审批语',
    information     text comment '审批对象信息',
    status          tinyint unsigned not null comment '状态，1=待处理、2=驳回、3=同意',
    created_time    timestamp comment '创建时间',
    finished_time   timestamp comment '结束时间',
    index idx_app_id_type (app_id, type) using btree,
    index idx_applier_id (applier_id) using btree,
    index idx_approver_id (approver_id) using btree,
    index idx_status (status) using btree,
    index idx_created_time (created_time) using btree
) engine innodb comment '待办表';

create table if not exists t_windows_certificate
(
    id                              int unsigned primary key auto_increment comment '自增主码',
    certificate_id                  char(32)         not null comment '唯一标识',
    app_id                          int unsigned comment '所属应用 ID',
    user_id                         int unsigned     not null comment '创建人 ID',
    sha1                            char(40)         not null comment '指纹',
    type                            tinyint unsigned not null comment '类型，1=公司EV、2=公司OV、3=个人EV、4=个人OV',
    ip                              varchar(15) comment 'UKEY 所在机器 IP',
    is_microsoft_verify_certificate bit(1) comment '是否是微软网站的验证证书',
    password                        varchar(64)      not null comment '密码',
    version                         int unsigned comment '版本号',
    publisher                       varchar(1024) comment '颁发者',
    common_name                     varchar(64)      not null comment '通用名',
    owner                           varchar(1024) comment '所有者',
    signature_algorithm             varchar(64) comment '签名算法',
    public_key_algorithm            varchar(64) comment '公钥密码算法',
    serial_number                   varchar(1024) comment '序列号',
    not_before                      datetime         not null comment '生效时间',
    not_after                       datetime         not null comment '失效时间',
    aes_key_id                      int unsigned comment '证书加密密钥 ID',
    content                         mediumblob comment '加密了的证书内容',
    created_time                    timestamp        not null comment '创建时间',
    deleted_time                    timestamp comment '删除时间',
    index idx_app_id (app_id) using btree,
    unique index uidx_certificate_id (certificate_id) using btree
) engine innodb comment 'Windows 证书信息表';

create table if not exists t_android_certificate
(
    id                   int unsigned primary key auto_increment comment '自增主码',
    certificate_id       char(32)         not null comment '唯一标识',
    app_id               int unsigned     not null comment '所属应用 ID',
    user_id              int unsigned     not null comment '创建人 ID',
    alias                varchar(64)      not null comment '别名',
    category             tinyint unsigned not null comment '类型，1=调试、2=发布',
    publisher            varchar(1024) comment '颁发者',
    owner                varchar(1024) comment '所有者',
    signature_algorithm  varchar(64) comment '签名算法',
    public_key_algorithm varchar(64) comment '公钥密码算法',
    version              varchar(32) comment '版本号',
    serial_number        varchar(64) comment '序列号',
    store_provider       varchar(64) comment '证书提供方',
    store_type           varchar(64) comment '证书格式类型',
    creation_date        date comment '证书创建时间',
    sha1                 char(40) comment '证书 SHA1 摘要值',
    sha256               char(64) comment '证书 SHA256 摘要值',
    storepass            varchar(32)      not null comment '密钥库解密密码',
    keypass              varchar(32)      not null comment '私钥解密密码',
    not_before           datetime         not null comment '生效时间',
    not_after            datetime         not null comment '失效时间',
    aes_key_id           int unsigned     not null comment '证书加密密钥 ID',
    content              mediumblob       not null comment '加密了的密钥库',
    created_time         timestamp        not null comment '创建时间',
    deleted_time         timestamp comment '删除时间',
    index idx_app_id (app_id) using btree,
    unique index uidx_certificate_id (certificate_id) using btree
) engine innodb comment '安卓证书信息表';

create table if not exists t_apple_certificate
(
    id                   int unsigned primary key auto_increment comment '自增主码',
    certificate_id       char(32)         not null comment '唯一标识',
    app_id               int unsigned comment 'Push 证书所属应用 ID',
    user_id              int unsigned     not null comment '创建人 ID',
    in_apple_id          char(10)         not null comment '在苹果服务器的 ID',
    environment          tinyint unsigned comment 'Push 证书环境，1=正式、2=开发、3=企业内测',
    bundle_id            int unsigned comment 'Push 证书对应的 BundleID',
    category             tinyint unsigned not null comment '文件类型，1=签名证书、2=Push 证书',
    type                 varchar(32) comment '签名证书类型，由苹果 API 中定义，例如：IOS_DEVELOPMENT',
    password             varchar(16)      not null comment '密码',
    publisher            varchar(1024) comment '颁发者',
    owner                varchar(1024) comment '所有者',
    signature_algorithm  varchar(64) comment '签名算法',
    public_key_algorithm varchar(64) comment '公钥密码算法',
    serial_number        varchar(64) comment '序列号',
    not_before           datetime         not null comment '生效时间',
    aes_key_id           int unsigned     not null comment '证书加密密钥 ID',
    not_after            datetime         not null comment '失效时间',
    content              mediumblob       not null comment '加密了的证书内容',
    created_time         timestamp        not null comment '创建时间',
    deleted_time         timestamp comment '删除时间',
    index idx_app_id (app_id) using btree,
    unique index uidx_certificate_id (certificate_id) using btree
) engine innodb comment '苹果相关证书信息表';

create table if not exists t_apple_profile
(
    id             int unsigned primary key auto_increment comment '自增主码',
    profile_id     char(32)     not null comment '唯一标识',
    app_id         int unsigned not null comment '所属应用 ID',
    user_id        int unsigned not null comment '创建人 ID',
    certificate_id int unsigned not null comment '证书表 ID，根据该证书申请的描述文件',
    in_apple_id    char(10)     not null comment '在苹果服务器的 ID',
    bundle_id      int unsigned comment '描述文件对应的 BundleID',
    text           text         not null comment '描述文件中的 XML 内容',
    type           varchar(32)  not null comment '描述文件类型，由苹果 API 中定义，例如：IOS_APP_STORE',
    not_before     datetime     not null comment '生效时间',
    not_after      datetime     not null comment '失效时间',
    content        mediumblob   not null comment '描述文件的内容',
    created_time   timestamp    not null comment '创建时间',
    deleted_time   timestamp comment '删除时间',
    index idx_app_id (app_id) using btree,
    unique index uidx_profile_id (profile_id) using btree
) engine innodb comment '苹果描述文件信息表';

create table if not exists t_apple_bundle_id
(
    id           int unsigned primary key auto_increment comment '自增主码',
    app_id       int unsigned     not null comment '所属应用 ID',
    user_id      int unsigned     not null comment '申请人 ID',
    in_apple_id  char(10)         not null comment '在苹果服务中的 ID',
    bundle_id    varchar(64)      not null comment 'Apple 应用唯一标识，域名倒置点号分隔的形式',
    environment  tinyint unsigned not null comment '环境，1=AppStore，2=企业内测',
    platform     tinyint unsigned not null comment '平台，1=IOS、2=MAC_OS、3=UNIVERSAL',
    capabilities mediumtext comment 'BundleID 的能力项，以英文逗号分隔',
    created_time timestamp        not null comment '创建时间',
    updated_time timestamp comment '更新时间',
    index idx_app_id (app_id) using btree,
    unique index uidx_bundle_id (bundle_id) using btree
) engine innodb comment 'BundleID 信息表';

create table if not exists t_apple_device
(
    id           int unsigned primary key auto_increment comment '自增主码',
    app_id       int unsigned     not null comment '所属应用 ID',
    user_id      int unsigned     not null comment '申请人 ID',
    in_apple_id  char(10) comment '在苹果服务器中的 ID',
    udid         varchar(128)     not null comment '设备 ID',
    model        varchar(32) comment '设备型号',
    remark       varchar(1024) comment '备注',
    platform     varchar(64)      not null comment '平台',
    status       tinyint unsigned not null comment '状态，1=正常、2=待审核、3=未通过',
    created_time timestamp        not null comment '创建时间',
    updated_time timestamp comment '更新时间',
    index idx_app_id (app_id) using btree
) engine innodb comment '苹果测试设备信息表';

create table if not exists t_api_account
(
    id           int unsigned primary key auto_increment comment '自增主码',
    app_id       int unsigned not null comment '所属应用 ID',
    user_id      int unsigned not null comment '申请人 ID',
    account_id   varchar(64)  not null comment '凭证 ID',
    ip           varchar(256) not null comment '凭证可以放行的 IP、IP 段，*号表示所有，逗号分隔',
    frequency    int unsigned not null comment '每分钟最高请求次数',
    secret       char(128)    not null comment '密钥',
    expired_time timestamp    not null comment '失效时间',
    created_time timestamp    not null comment '创建时间',
    updated_time timestamp comment '创建时间',
    unique index idx_app_id_account_id (app_id, account_id) using btree
) engine innodb comment 'OpenAPI 凭证信息表';

create table if not exists t_api_authorization
(
    id             int unsigned primary key auto_increment comment '自增主码',
    api_account_id int unsigned comment 'OpenAPI 凭证信息表 ID',
    capability     tinyint unsigned not null comment '权限项，对应表 3-5',
    index idx_api_account_id (api_account_id) using btree
) engine innodb comment 'OpenAPI 凭证授权表';

create table if not exists t_windows_signing_job
(
    id             int unsigned primary key auto_increment comment '自增主码',
    job_id         char(38)         not null comment '任务 ID，前 6 为是年月，后面是 UUID',
    type           tinyint unsigned not null comment '类型，1=PE 签名、2=微软 Attestation 签名、3=PE+微软 Attestation 签名、4=HLKX 文件签名',
    app_id         int unsigned comment '所属应用 ID',
    user_id        int unsigned     not null comment '提交人 ID',
    certificate_id int unsigned     not null comment 'Windows 证书表 ID',
    file_id        char(38)         not null comment '待签名文件 ID',
    signed_file_id char(38) comment '已签名文件 ID',
    log            mediumtext comment '任务日志',
    product_id     char(17) comment '微软侧的 ID',
    submission_id  char(19) comment '微软侧的 ID',
    source         tinyint unsigned not null comment '来源，1=页面，2=OpenAPI、3=内部',
    finished_time  timestamp comment '结束时间',
    finish_pe_time timestamp comment '完成 PE 签名时间',
    status         tinyint unsigned not null comment '状态，1=签名中、2=等待 Cab 签名、3=Cab 签名中、4=待 Attestation 签名、5=Attestation 签名中、6=失败、7=成功',
    created_time   timestamp        not null comment '提交时间',
    updated_time   timestamp on update current_timestamp() comment '更新时间',
    index idx_app_id (app_id) using btree,
    index idx_status (status) using btree,
    index idx_created_time (created_time) using btree,
    unique index uidx_job_id (job_id) using btree
) engine innodb comment 'Windows 签名任务表';

create table if not exists t_whql_job
(
    id                int unsigned primary key auto_increment comment '自增主码',
    job_id            char(32)         not null comment '任务 ID',
    app_id            int unsigned     not null comment '所属应用 ID',
    user_id           int unsigned     not null comment '提交人 ID',
    file_id           char(38)         not null comment '待签名文件 ID',
    signed_file_id    char(38) comment '已签名文件 ID',
    hlk_log_file_id   char(38) comment 'HLK Studio 日志包文件 ID',
    hlkx_file_id      char(38) comment 'HLKX 文件 ID',
    hlkx_sign_job_id  char(38) comment 'HLKX 文件签名任务 ID，对应 Windows 任务表中的 ID',
    type              tinyint unsigned not null comment '类型，1=HLK 兼容性测试+WHQL 签名、2=仅 WHQL 签名',
    log               mediumtext comment '任务日志',
    product_id        char(17) comment '微软侧的 ID',
    submission_id     char(19) comment '微软侧的 ID',
    source            tinyint unsigned not null comment '来源，1=页面、2=OpenAPI',
    test_system       varchar(32) comment '测试系统',
    test_machine_name char(15) comment '测试机器名',
    test_target       varchar(256) comment '测试目标名',
    service_name      varchar(256) comment '服务名',
    test_config       text comment 'HLK 测试项配置',
    finished_time     timestamp comment '结束时间',
    finish_test_time  timestamp comment '测试结束时间',
    status            tinyint unsigned not null comment '状态，1=待HLK测试、2=测试机初始化中、3=测试机初始化完毕，4=启动HLK测试中，5=HLK测试中，6=HLK测试完毕、7=HLKX 文件签名中、8=等待 WHQL认证结果中、9=失败、10=成功',
    created_time      timestamp        not null comment '提交时间',
    updated_time      timestamp on update current_timestamp() comment '更新时间',
    index idx_app_id (app_id) using btree,
    index idx_status (status) using btree,
    index idx_created_time (created_time) using btree,
    unique index uidx_job_id (job_id) using btree
) engine innodb comment 'WHQL 签名任务表';

create table if not exists t_android_signing_job
(
    id                int unsigned primary key auto_increment comment '自增主码',
    job_id            char(38)         not null comment '任务 ID，前 6 为是年月，后面是 UUID',
    app_id            int unsigned     not null comment '所属应用 ID',
    user_id           int unsigned     not null comment '提交人 ID',
    type              tinyint unsigned not null comment '类型，1=APK 签名、2=AAB 签名、3=补丁签名',
    certificate_id    int unsigned     not null comment '安卓证书表 ID',
    file_id           char(38)         not null comment '待签名文件 ID',
    signed_file_id    char(38) comment '已签名文件 ID',
    log               mediumtext comment '任务日志',
    signature_schemas varchar(32) comment 'APK 签名方案，英文逗号分隔',
    minimum_sdk_level int unsigned comment '补丁包签名 SDK API 最低版本号',
    source            tinyint unsigned not null comment '来源，1=页面、2=OpenAPI',
    finished_time     timestamp comment '结束时间',
    status            tinyint unsigned not null comment '状态，1=进行中，2=成功，3=失败',
    created_time      timestamp        not null comment '提交时间',
    index idx_app_id (app_id) using btree,
    index idx_created_time (created_time) using btree,
    unique index uidx_job_id (job_id) using btree
) engine innodb comment '安卓签名任务表';

create table if not exists t_apple_signing_job
(
    id             int unsigned primary key auto_increment comment '自增主码',
    job_id         char(38)         not null comment '任务 ID，前 6 为是年月，后面是 UUID',
    app_id         int unsigned     not null comment '所属应用 ID',
    user_id        int unsigned     not null comment '提交人 ID',
    profile_id     int unsigned     not null comment '描述文件表 ID',
    file_id        char(38)         not null comment '待签名文件 ID',
    signed_file_id char(38) comment '已签名文件 ID',
    log            mediumtext comment '任务日志',
    source         tinyint unsigned not null comment '来源，1=页面、2=OpenAPI',
    finished_time  timestamp comment '结束时间',
    status         tinyint unsigned not null comment '状态，1=进行中、2=成功、3=失败',
    created_time   timestamp        not null comment '提交时间',
    index idx_app_id (app_id) using btree,
    index idx_created_time (created_time) using btree,
    unique index uidx_job_id (job_id) using btree
) engine innodb comment '苹果签名任务表';

create table if not exists t_file
(
    id             int unsigned primary key auto_increment comment '自增主码',
    file_id        char(38)         not null comment '文件 ID，前 6 位是年月，后 32 位是 UUID',
    tusd_id        char(32)         not null comment '在 Tusd 服务器的 ID',
    user_id        int unsigned comment '上传人 ID',
    api_account_id int unsigned comment '上传凭证 ID',
    app_id         int unsigned comment '所属应用 ID',
    name           varchar(256) comment '文件名',
    md5            char(32) comment '文件摘要',
    size           int unsigned comment '文件大小',
    type           tinyint unsigned not null comment '类型；1=用户头像、2=应用图标、3=Android 签名文件、4=Windows 签名文件、5=Apple 签名文件、6=HLK 日志文件、7=微软方的结果文件',
    created_time   timestamp        not null comment '上传时间',
    index idx_md5 (md5) using btree,
    unique index idx_file_id (file_id) using btree
) engine innodb comment '文件信息表';

create table if not exists t_windows_certificate_authorization
(
    id             int unsigned primary key auto_increment comment '自增主码',
    app_id         int unsigned not null comment '所属应用 ID',
    user_id        int unsigned not null comment '操作人 ID',
    certificate_id int unsigned not null comment 'Windows 证书表 ID',
    created_time   timestamp    not null comment '授权时间',
    unique index idx_certificate_id_app_id (certificate_id, app_id) using btree
) engine innodb comment '授权应用 EV 证书表';

create table if not exists t_android_organization
(
    id           int unsigned primary key auto_increment comment '自增主码',
    name         varchar(32)  not null unique comment '通用名 CN',
    user_id      int unsigned not null comment '添加人 ID',
    owner        varchar(256) not null unique comment '所有者，用于 keytool 中的 dname 参数值',
    created_time timestamp    not null comment '创建时间'
) engine innodb comment '安卓证书主体表';

create table if not exists t_notice
(
    id             int unsigned primary key auto_increment comment '自增主码',
    content        mediumtext   not null comment '公告内容',
    user_id        int unsigned not null comment '创建人 ID',
    expired_time   timestamp    not null comment '失效时间',
    activated_time timestamp    not null comment '生效时间',
    created_time   timestamp    not null comment '创建时间',
    index idx_created_time (created_time) using btree
) engine innodb comment '公告信息表';

create table if not exists t_aes_key
(
    id           int unsigned primary key auto_increment comment '自增主码',
    secret       binary(16) check (length(secret) = 16) not null comment '密钥',
    created_time timestamp                              not null default current_timestamp() comment '创建时间'
) engine innodb comment '证书加密密钥表';

insert ignore into t_user (id, name_en, name_zh, avatar_file_id, department, password_digest, password_salt,
                           created_time)
values (1, 'admin', '系统管理员', '', '', '21232f297a57a5a743894a0e4a801fc3', '', now());

insert ignore into t_user_role (user_id, role)
values (1, 1);
