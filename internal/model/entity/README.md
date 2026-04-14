# Entity 表结构与关系说明

## 表清单

### 主表（8 张）

| 表名 | 说明 | 树形 |
|------|------|------|
| `sys_user` | 系统管理员 | - |
| `sys_role` | 系统角色（支持继承） | `ParentID` 树形 |
| `sys_menu` | 系统菜单/按钮权限 | `ParentID` 树形 |
| `sys_dept` | 系统部门 | `ParentID` 树形 |
| `sys_post` | 系统岗位 | - |
| `sys_api` | API 接口定义 | - |
| `sys_api_group` | API 分组（支持嵌套） | `ParentID` 树形 |
| `sys_config` | 系统配置（分组+配置项自引用） | `ParentID` 自引用 |

### 关联表（5 张）

| 表名 | 说明 | 关联方向 |
|------|------|----------|
| `sys_user_role` | 用户-角色 | `sys_user` ↔ `sys_role` |
| `sys_user_post` | 用户-岗位 | `sys_user` ↔ `sys_post` |
| `sys_role_menu` | 角色-菜单 | `sys_role` ↔ `sys_menu` |
| `sys_role_dept` | 角色-部门（数据权限） | `sys_role` ↔ `sys_dept` |
| `sys_menu_api` | 菜单-API（按钮权限） | `sys_menu` ↔ `sys_api` |

## ER 关系图

```
                          ┌─────────────┐
                          │  sys_config  │（自引用: 分组行 → 配置项行）
                          └─────────────┘

┌──────────┐    N:N    ┌──────────┐    N:N    ┌──────────┐
│ sys_user │◄─────────►│ sys_role │◄─────────►│ sys_menu │
└────┬─────┘           └────┬─────┘           └────┬─────┘
     │                      │                      │
     │ N:N                  │ N:N                  │ N:N
     ▼                      ▼                      ▼
┌──────────┐          ┌──────────┐           ┌──────────┐
│ sys_post │          │ sys_dept │           │  sys_api │
└──────────┘          └──────────┘           └────┬─────┘
                                                  │
                                             ┌────┴─────┐
                                             │sys_api_   │
                                             │  group    │
                                             └──────────┘

关联表:
  sys_user_role   → user_id  ↔ role_id
  sys_user_post   → user_id  ↔ post_id
  sys_role_menu   → role_id  ↔ menu_id
  sys_role_dept   → role_id  ↔ dept_id
  sys_menu_api    → menu_id  ↔ api_id

外键归属:
  sys_user.dept_id → sys_dept.id （多对一）
```

## 核心业务关系

### 用户与权限

```
用户(sys_user)
 ├── 所属部门 → sys_dept（dept_id）
 ├── 担任岗位 → sys_post（通过 sys_user_post）
 └── 拥有角色 → sys_role（通过 sys_user_role）
       ├── 菜单权限 → sys_menu（通过 sys_role_menu）
       │     └── API 权限 → sys_api（通过 sys_menu_api）
       └── 数据权限 → sys_dept（通过 sys_role_dept，仅 DataScope=2 时）
```

### 权限校验链路

```
1. 菜单权限（前端按钮/页面显示）:
   用户 → 角色 → 菜单（sys_role_menu）→ 菜单树

2. API 权限（后端接口鉴权，由 Casbin 管理）:
   Casbin 策略(role, path, method) → 放行/拒绝

3. 数据权限（行级数据隔离）:
   角色.DataScope 决定可见范围:
     1=全部数据
     2=自定义（通过 sys_role_dept 指定部门）
     3=本部门
     4=本部门及下级
     5=仅本人
```

### 树形结构

4 张树形表统一使用 `ParentID(*int64)` + `Level` + `Tree(物化路径)` 方案：

| 表 | ParentID | Level | Tree | 说明 |
|---|---|---|---|---|
| `sys_dept` | `*int64` | int | varchar(255) | 部门层级 |
| `sys_menu` | `*int64` | int | varchar(255) | 菜单目录/页面/按钮 |
| `sys_role` | `*int64` | int | varchar(255) | 角色继承 |
| `sys_api_group` | `*int64` | int | varchar(255) | API 分组嵌套 |

| 字段 | 说明 | 示例 |
|------|------|------|
| `ParentID` | 父节点 ID，`nil` 表示顶级 | `nil` |
| `Level` | 层级深度，顶级为 0 | `0` |
| `Tree` | 物化路径，逗号分隔祖先 ID | `"1,3,7"` |

查询子树：`WHERE tree LIKE '1,3,%'` 或递归 `ParentID`。

### 配置表结构（sys_config 自引用）

单表实现分组 + 配置项，通过 `ParentID` 区分两种角色：

```
分组行 (ParentID=nil):
  ConfigKey="email"  Name="邮箱配置"  Icon="mail"
    ├── ConfigKey="host"     Name="SMTP地址"  ConfigValue="smtp.ex.com"  ValueType="string"
    ├── ConfigKey="port"     Name="SMTP端口"  ConfigValue="465"          ValueType="number"
    └── ConfigKey="password" Name="SMTP密码"  ConfigValue="***"          ValueType="string"  IsEncrypted=true
```

| 角色 | 使用字段 | 空字段 |
|------|---------|--------|
| 分组行 | ConfigKey, Name, Icon, Description, Sort, Status | ConfigValue, ValueType, IsEncrypted |
| 配置项行 | ConfigKey, Name, ConfigValue, ValueType, IsEncrypted, Sort, Status | Icon |

唯一约束：`(parent_id, config_key)` — 同分组内键名唯一。
GORM 关联：`Children []*SysConfig` — Preload 一次加载分组 + 全部子项。

## 待补充

| 表名 | 说明 | 优先级 |
|------|------|--------|
| `sys_audit_log` | 操作审计日志（记录增删改操作、请求参数、耗时） | 高 |
| `sys_login_log` | 登录日志（记录登录/登出/IP/设备/位置） | 中 |
| `sys_dict_type` | 字典类型（通用下拉、状态枚举管理） | 中 |
| `sys_dict_data` | 字典数据（字典项键值对） | 中 |

## 字段约定

| 约定 | 说明 |
|------|------|
| 主键 | `id bigint auto_increment` |
| 软删除 | `deleted_at timestamp(3)`，主表都有，关联表无 |
| 审计字段 | `created_by` / `updated_by`，记录操作人 ID |
| 时间精度 | `timestamp(3)` 毫秒级 |
| 状态字段 | `status smallint`，1=启用 0=禁用 |
| 排序字段 | `sort int`，升序 |
| 内置标记 | `builtin boolean`，内置数据不可删除 |
