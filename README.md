# PoorSquad

:call_me_hand: 穷逼小分队，GitHub 账号管理面板。转为中小型团队、工作室在 GitHub 愉快协作管理雇员使用。

## 产品文档

### 基本单位

- 用户
  - 超级管理员：第一个登录到系统的人，具有最高（所有）权限
  - 企业：每个用户可以自由添加企业
    - 企业管理员：管理企业绑定的账号、管理企业团队、绑定团队项目、项目的所有设置。具有该企业最高（所有）权限
    - 企业成员：可以查看企业信息
    - 绑定的 GitHub 账号
    - 项目
      - 外部贡献者：单项目外部贡献者，只能阅读单项目内信息
      - branch：保护分支、删除分支
      - webhook：添加修改删除触发 webhook
    - 小组
      - 小组管理员：具有管理项目设置（Webhook、Protect Branch、Deploy Key）、项目下成员的权限
      - 小组成员：读取项目的所有信息、触发 webhook
  