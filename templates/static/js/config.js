// 系统API配置
const API_CONFIG = {
  // 基础路径
  BASE_URL: 'http://localhost:9080',
  
  // 登录类型
  LOGIN_TYPE: {
    USERNAME: 'username',   // 用户名密码登录
    MOBILE: 'mobile',       // 手机号验证码登录
    EMAIL: 'email',         // 邮箱验证码登录
    WECHAT_WORK: 'wechat_work', // 企业微信登录
  },
  
  // 管理员相关API
  ADMIN: {
    // 用户管理
    USER: {
      LOGIN: '/admin/user/login',           // 用户登录
      LOGOUT: '/admin/user/logout',         // 用户退出
      SEND_CODE: '/admin/user/send-code',   // 发送验证码
      OAUTH_INIT: '/admin/user/oauth-init', // OAuth初始化

      SET_PASSWORD: '/admin/user/set-password',          // 设置密码
      CHANGE_PASSWORD: '/admin/user/change-password',    // 修改密码
      BIND_MOBILE: '/admin/user/bind-mobile',            // 绑定手机号
      UNBIND_MOBILE: '/admin/user/unbind-mobile',        // 解绑手机号

      LIST: '/admin/user/list',                          // 获取用户列表
      UPDATE_STATUS: '/admin/user/update-status',        // 更新用户状态


      CURRENT_INFO: '/admin/user/current-info',          // 获取当前用户信息
      UPDATE_PROFILE: '/admin/user/update-profile',      // 更新当前用户个人信息
      MENUS_BUTTONS: '/admin/user/menus-buttons',         // 获取用户菜单和按钮权限
      EDIT: '/admin/tpl/page/account/user-detail.html',  // 编辑用户
      UPDATE : '/admin/user/update',                     // 更新用户
      INFO: '/admin/user/info',                          // 获取用户信息
      DELETE: '/admin/user/delete',                      // 删除用户
      ADD: '/admin/tpl/page/account/user-add.html',      // 新增用户页面路径
      USER_LIST: '/admin/tpl/page/account/user.html',    // 用户列表
      ADD_USER: '/admin/user/add',                       // 新增用户
    },

    ROLE: {
      GET_USER_ROLE_PERMISSIONS: '/admin/role/get-user-role-permissions', // 获取用户角色和权限, 用户查看详情页
      LIST: '/admin/role/list',                           // 获取角色列表
      INFO: '/admin/role/detail',                         // 获取角色详情, 用户查看角色详情页
      UPDATE_STATUS: '/admin/role/update-status',         // 更新角色状态
      DELETE: '/admin/role/delete',                       // 删除角色
      EDIT_INFO: '/admin/role/edit-info',                 // 获取角色信息用于编辑
      UPDATE: '/admin/role/update',                       // 更新角色相关信息，包括角色名称、权限等


      ROLE_LIST: '/admin/tpl/page/account/role.html',     // 角色列表
      EDIT: '/admin/tpl/page/account/role-edit.html',     // 编辑角色页，用于编辑角色
      DETAIL: '/admin/tpl/page/account/role-detail.html', // 角色详情页, 用于查看
      UPDATE_IS_SYSTEM: '/admin/role/update-is-system',   // 更新是否系统角色

      ADD: '/admin/tpl/page/account/role-add.html',       // 新增角色页
      ADD_ROLE: '/admin/role/add',                        // 新增角色
      GET_ALL_PERMISSIONS: '/admin/role/get-all-permissions', // 获取系统所有权限点
    },

    DEPT: {
      LIST: '/admin/dept/list',                           // 获取部门列表
      INFO: '/admin/dept/detail',                         // 获取部门详情
      UPDATE_STATUS: '/admin/dept/update-status',         // 更新部门状态
      DELETE: '/admin/dept/delete',                       // 删除部门
      EDIT_INFO: '/admin/dept/edit-info',                 // 获取部门信息用于编辑
      UPDATE: '/admin/dept/update',                       // 更新部门相关信息
      ADD: '/admin/tpl/page/department/dept-add.html',       // 新增部门页
      ADD_DEPT: '/admin/dept/add',                        // 新增部门
      DEPT_LIST: '/admin/tpl/page/department/dept.html',     // 部门列表
      EDIT: '/admin/tpl/page/department/dept-edit.html',     // 编辑部门页
      DETAIL: '/admin/tpl/page/department/dept-detail.html', // 部门详情页
      ADD_DEPT_USER: '/admin/tpl/page/department/dept-user-add.html', // 添加部门员工页
      USER_LIST: '/admin/dept/user-list',                 // 获取部门员工列表
      ADD_USER: '/admin/dept/add-user',                   // 添加部门员工
      UPDATE_USER: '/admin/dept/update-user',             // 更新部门员工
      REMOVE_USER: '/admin/dept/remove-user',            // 移除部门员工
      BATCH_UPDATE_USER: '/admin/dept/batch-update-user', // 批量更新部门员工
      GET_ALL: '/admin/dept/get-all',                     // 获取所有部门列表（用于下拉框）
    },

    PERMISSION: {
      LIST: '/admin/permission/list',                     // 获取权限列表
      ADD_PERMISSION: '/admin/permission/add',            // 新增权限
      UPDATE: '/admin/permission/update',                 // 更新权限
      UPDATE_IS_SYSTEM: '/admin/permission/update-is-system', // 更新是否系统权限
      DELETE: '/admin/permission/delete',                 // 删除权限
      UPDATE_STATUS: '/admin/permission/update-status',   // 更新权限状态
      EDIT_INFO: '/admin/permission/edit-info',           // 获取编辑权限信息
      GET_TREE: '/admin/permission/get-tree',             // 获取权限树

      PERMISSION_LIST: '/admin/tpl/page/account/permission.html', // 权限列表
      ADD: '/admin/tpl/page/account/permission-add.html', // 新增权限页
      EDIT: '/admin/tpl/page/account/permission-add.html', // 编辑权限页（与新增共用一个页面）
    },

    LOGIN_LOG: {
      LIST: '/admin/login-log/list',               // 获取登录日志列表
    },

  },
  
  // 页面路由
  PAGE: {
    HOME: '/admin/tpl/home.html',                // 首页
    ADMIN_LOGIN: '/admin/tpl/',              // 管理员登录页
  },
  
  // 静态资源
  STATIC: {
    CSS: '/static/css/',
    JS: '/static/js/',
    IMAGES: '/static/images/',
  },
  
  // 文件下载
  DOWNLOADS: '/downloads/',
};

// 辅助函数：构建完整URL
API_CONFIG.buildURL = function(path) {
  return this.BASE_URL + path;
};

