window.app = {
    // 最标准的函数定义方式
    ajax(options) {
        const defaults = {
          type: 'POST',
          url: '',
          data: {},
          dataType: 'json',
          timeout: 30000,
          contentType: 'application/x-www-form-urlencoded',
          headers: {
            'x-csrf-token': Cookies.get('x-csrf-token'),
          },
          xhrFields: {
            withCredentials: true, // 允许跨域请求携带cookie
          },
          success: null,
          error: null
        };
        // console.log("x-csrf-token: ", Cookies.get('x-csrf-token'));
        const settings = Object.assign({}, defaults, options);
        try {
          $.ajax({
            type: settings.type,
            url: settings.url,
            data: settings.data,
            dataType: settings.dataType,
            timeout: settings.timeout,
            contentType: settings.contentType,
            headers: settings.headers,
            xhrFields: settings.xhrFields,
            success(data) {
              if (data.code == 0) {
                if (typeof settings.success === 'function') {
                  settings.success(data);
                }
              } else if (data.code == 302) {
                if (window === window.top) tools.jumpUrl(data.data.url);
                else NUI.postMessage({ type: 'jumpUrl', data: data.data.url });
              } else if (data.code == 401) {
                // 401 未授权，跳转登录页
                console.log(data);
                tools.showMsg(data.msg, 0);
                tools.jumpUrl(API_CONFIG.PAGE.ADMIN_LOGIN, 'logout=1');
              } else {
                tools.showMsg(data.msg, 0);
                tools.checkErrCode(data.code);
                if (typeof settings.success === 'function') {
                  settings.success(data);
                }
              }
            },
            error(error) {
              tools.log(error);
              if (typeof settings.error === 'function') {
                settings.error(error);
              } else {
                if (error.status == 404 || error.status == 500) {
                  tools.showMsg(options.error_msg || '请求失败', 0);
                }
                if (typeof settings.success === 'function') {
                  settings.success();
                }
              }
            }
          });
        } catch (e) {
          tools.log(e);
          if (typeof settings.error === 'function') {
            settings.error(e);
          } else if (typeof settings.success === 'function') {
            settings.success();
          }
        } finally {
        }
    },

    postJson(options) {
      const defaults = {
        url: '',
        data: {},
        success: null,
        error: null
      }
      const settings = Object.assign({}, defaults, options);
      try {
        this.ajax({
            type: 'POST',
            url: settings.url,
            data: JSON.stringify(settings.data),
            dataType: 'json',
            timeout: 30000,
            contentType: 'application/json',
            success: settings.success,
            error: settings.error
        });
      } catch (e) {
        tools.log(e);
        if (typeof settings.error === 'function') {
          settings.error(e);
        } else if (typeof settings.success === 'function') {
          settings.success();
        }
      } finally {
      }
    },

    postForm(options) {
      const defaults = {
        url: '',
        data: {},
        success: null,
        error: null
      }
      const settings = Object.assign({}, defaults, options);
      try {
        this.ajax({
            type: 'POST',
            url: settings.url,
            data: settings.data,
            dataType: 'json',
            timeout: 30000,
            contentType: 'application/x-www-form-urlencoded',
            success: settings.success,
            error: settings.error
        });
      } catch (e) {
        tools.log(e);
        if (typeof settings.error === 'function') {
          settings.error(e);
        } else if (typeof settings.success === 'function') {
          settings.success();
        }
      } finally {
      }
    },

    /**
     * 校验用户名
     * @param {String} username 用户名
     * @returns {Object} {valid: boolean, message: string} 校验结果
     */
    validateUsername(username) {
      if (!username || username.trim() === '') {
        return { valid: false, message: '用户名不能为空' };
      }
      const trimmedUsername = username.trim();
      if (trimmedUsername.length < 5 || trimmedUsername.length > 20) {
        return { valid: false, message: '用户名长度必须在5-20位之间' };
      }
      // 用户名只能是字母、数字、英文符号（允许的英文符号：_ - . @）
      const usernamePattern = /^[a-zA-Z0-9_\-\.@]+$/;
      if (!usernamePattern.test(trimmedUsername)) {
        return { valid: false, message: '用户名只能包含字母、数字和英文符号（_、-、.、@）' };
      }
      return { valid: true, message: '' };
    },

    /**
     * 校验密码
     * @param {String} password 密码
     * @returns {Object} {valid: boolean, message: string} 校验结果
     */
    validatePassword(password) {
      if (!password || password.trim() === '') {
        return { valid: false, message: '密码不能为空' };
      }
      const trimmedPassword = password;
      if (trimmedPassword.length < 6 || trimmedPassword.length > 20) {
        return { valid: false, message: '密码长度必须在6-20位之间' };
      }
      // 密码只能包含字母、数字、英文符号
      const passwordPattern = /^[a-zA-Z0-9_\-\.@!#$%^&*()]+$/;
      if (!passwordPattern.test(trimmedPassword)) {
        return { valid: false, message: '密码只能包含字母、数字和英文符号' };
      }
      return { valid: true, message: '' };
    },

    /**
     * 校验手机号
     * @param {String} mobile 手机号
     * @returns {Object} {valid: boolean, message: string} 校验结果
     */
    validateMobile(mobile) {
      if (!mobile || mobile.trim() === '') {
        return { valid: false, message: '手机号不能为空' };
      }
      const trimmedMobile = mobile.trim();
      // 手机号必须是11位数字，且以1开头（1开头，第二位3-9，后面9位数字）
      const mobilePattern = /^1[3-9]\d{9}$/;
      if (!mobilePattern.test(trimmedMobile)) {
        return { valid: false, message: '手机号格式不正确，必须是11位数字且以1开头' };
      }
      return { valid: true, message: '' };
    },

    /**
     * 检查用户是否有指定按钮权限
     * @param {string} permissionCode - 按钮权限编码，如 'system.user.add'
     * @returns {boolean} - 是否有权限
     */
    hasButtonPermission(permissionCode) {
      if (!permissionCode) {
        return false;
      }
      // 优先从NUI.data中获取（home.html中）
      if (window.NUI && window.NUI.data && window.NUI.data.buttonPermissions) {
        return window.NUI.data.buttonPermissions.indexOf(permissionCode) !== -1;
      }
      // 从localStorage中获取（其他页面）
      try {
        var buttonPermissions = JSON.parse(localStorage.getItem('BUTTON_PERMISSIONS') || '[]');
        return buttonPermissions.indexOf(permissionCode) !== -1;
      } catch (e) {
        console.error('获取按钮权限失败:', e);
        return false;
      }
    },

    /**
     * 校验真实姓名
     * @param {String} realname 真实姓名
     * @returns {Object} {valid: boolean, message: string} 校验结果
     */
    validateRealname(realname) {
      // 真实姓名可以为空，如果不为空则校验格式
      if (!realname || realname.trim() === '') {
        return { valid: true, message: '' };
      }
      const trimmedRealname = realname.trim();
      // 真实姓名长度限制：2-20个字符
      if (trimmedRealname.length < 2 || trimmedRealname.length > 20) {
        return { valid: false, message: '真实姓名长度必须在2-20个字符之间' };
      }
      // 真实姓名可以包含中文、字母、数字、空格、点号、连字符
      const realnamePattern = /^[\u4e00-\u9fa5a-zA-Z0-9\s\.\-]+$/;
      if (!realnamePattern.test(trimmedRealname)) {
        return { valid: false, message: '真实姓名只能包含中文、字母、数字、空格、点号和连字符' };
      }
      return { valid: true, message: '' };
    },

    /**
     * 校验昵称
     * @param {String} nickname 昵称
     * @returns {Object} {valid: boolean, message: string} 校验结果
     */
    validateNickname(nickname) {
      // 昵称可以为空，如果不为空则校验格式
      if (!nickname || nickname.trim() === '') {
        return { valid: true, message: '' };
      }
      const trimmedNickname = nickname.trim();
      // 昵称长度限制：1-30个字符
      if (trimmedNickname.length < 1 || trimmedNickname.length > 30) {
        return { valid: false, message: '昵称长度必须在1-30个字符之间' };
      }
      // 昵称可以包含中文、字母、数字、下划线、连字符、点号、空格
      const nicknamePattern = /^[\u4e00-\u9fa5a-zA-Z0-9_\-\s\.]+$/;
      if (!nicknamePattern.test(trimmedNickname)) {
        return { valid: false, message: '昵称只能包含中文、字母、数字、下划线、连字符、点号和空格' };
      }
      return { valid: true, message: '' };
    },

    /**
     * 校验邮箱
     * @param {String} email 邮箱
     * @returns {Object} {valid: boolean, message: string} 校验结果
     */
    validateEmail(email) {
      // 邮箱可以为空，如果不为空则校验格式
      if (!email || email.trim() === '') {
        return { valid: true, message: '' };
      }
      const trimmedEmail = email.trim();
      // 邮箱长度限制：5-100个字符
      if (trimmedEmail.length < 5 || trimmedEmail.length > 100) {
        return { valid: false, message: '邮箱长度必须在5-100个字符之间' };
      }
      // 邮箱格式校验：标准邮箱格式
      const emailPattern = /^[a-zA-Z0-9][a-zA-Z0-9._-]*@[a-zA-Z0-9][a-zA-Z0-9.-]*\.[a-zA-Z]{2,}$/;
      if (!emailPattern.test(trimmedEmail)) {
        return { valid: false, message: '邮箱格式不正确' };
      }
      return { valid: true, message: '' };
    },

    /**
     * 校验生日
     * @param {String} birthday 生日，格式：YYYY-MM-DD 或 YYYY/MM/DD
     * @returns {Object} {valid: boolean, message: string} 校验结果
     */
    validateBirthday(birthday) {
      // 生日可以为空，如果不为空则校验格式
      if (!birthday || birthday.trim() === '') {
        return { valid: true, message: '' };
      }
      const trimmedBirthday = birthday.trim();
      // 生日格式校验：YYYY-MM-DD 或 YYYY/MM/DD
      const birthdayPattern = /^(\d{4})[-/](\d{1,2})[-/](\d{1,2})$/;
      if (!birthdayPattern.test(trimmedBirthday)) {
        return { valid: false, message: '生日格式不正确，请使用 YYYY-MM-DD 或 YYYY/MM/DD 格式' };
      }
      // 验证日期是否有效
      const parts = trimmedBirthday.split(/[-/]/);
      const year = parseInt(parts[0], 10);
      const month = parseInt(parts[1], 10);
      const day = parseInt(parts[2], 10);
      
      // 年份范围：1900-当前年份
      const currentYear = new Date().getFullYear();
      if (year < 1900 || year > currentYear) {
        return { valid: false, message: '年份必须在1900-' + currentYear + '之间' };
      }
      
      // 月份范围：1-12
      if (month < 1 || month > 12) {
        return { valid: false, message: '月份必须在1-12之间' };
      }
      
      // 日期范围：根据月份和年份判断
      const daysInMonth = new Date(year, month, 0).getDate();
      if (day < 1 || day > daysInMonth) {
        return { valid: false, message: '日期无效' };
      }
      
      // 验证日期不能是未来日期
      const date = new Date(year, month - 1, day);
      const today = new Date();
      today.setHours(0, 0, 0, 0);
      if (date > today) {
        return { valid: false, message: '生日不能是未来日期' };
      }
      
      return { valid: true, message: '' };
    },
};

