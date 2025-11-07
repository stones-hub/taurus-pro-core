; (function (window, document) {
  function bytesToSize(bytes, fractionDigits) {
    if (bytes === 0) return '0B'
    var k = 1024,
      sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'],
      i = Math.floor(Math.log(bytes) / Math.log(k))

    return (bytes / Math.pow(k, i)).toFixed(fractionDigits || 0) + sizes[i]
  }

  function cloneDeep(obj) {
    var type = Object.prototype.toString.call(obj).slice(8, -1)
    var result = type == 'Object' ? {} : type == 'Array' ? [] : obj
    for (var key in obj) {
      result[key] = typeof obj[key] == 'object' ? cloneDeep(obj[key]) : obj[key]
    }
    return result
  }

  // NUI.data 中带KEEP_前缀的属性会保存
  var keepRegexp = /^KEEP_/

  function getKeepKey() {
    return window.location.href.replace(/(\&|\?)keep=1/, '')
  }

  function saveRecordToSession() {
    var record = {},
      flag = false
    for (var key in NUI.data) {
      if (keepRegexp.test(key)) {
        flag = true
        record[key] = NUI.data[key]
      }
    }
    if (flag) {
      window.sessionStorage.setItem(getKeepKey(), JSON.stringify(record))
    }
  }

  // 当页面链接有keep=1参数时复原数据
  function recoverRecordFromSession(app) {
    // 页面全局刷新时清除页面数据记录session
    if (window.top === window) {
      window.sessionStorage.clear()
    } else {
      if (/keep=1/.test(window.location.search)) {
        var data = window.sessionStorage.getItem(getKeepKey())
        if (data) {
          var record = JSON.parse(data) || {}
          for (var key in record) {
            app[key] = record[key]
          }
        }
      }
    }
  }

  // 避免有全屏弹窗时，刷新iframe还保留父级遮罩
  window.addEventListener('beforeunload', function () {
    NUI.postMessage({ type: 'closeMask' })
    saveRecordToSession()
  })

  if (NUI.__app__) {
    recoverRecordFromSession(NUI.__app__)
  } else {
    NUI.EventHub.$on('nui:mounted', recoverRecordFromSession)
  }

  window.tools = {
    //判断是否微信浏览器
    isWeixinBrowser: function () {
      var ua = window.navigator.userAgent.toLowerCase()
      if (ua.match(/MicroMessenger/i) == 'micromessenger') {
        return true
      } else {
        return false
      }
    },
    // JS日志
    log: function (txt) {
      return console.log(txt)
    },
    // ajax请求
    ajax: function (
      type,
      url,
      data,
      callback,
      error_msg,
      timeout,
      contentType
    ) {
      try {
        $.ajax({
          type: type ? type : 'POST',
          url: url ? url : '',
          data: data ? data : {},
          dataType: 'json',
          timeout: timeout || 30000,
          contentType: contentType || 'application/x-www-form-urlencoded',
          success: function (data) {
            if (data.code == 0) {
              callback(data.data)
            } else if (data.code == 302) {
              if (window === window.top) tools.jumpUrl(data.data.url)
              else NUI.postMessage({ type: 'jumpUrl', data: data.data.url })
            } else {
              tools.showMsg(data.msg, 0)
              tools.checkErrCode(data.code)
              callback()
            }
          },
          error: function (error) {
            tools.log(error)
            if (error.status == 404 || error.status == 500) {
              if (typeof error_msg == 'function') {
                error_msg(error)
              } else {
                tools.showMsg(error_msg || '请求失败', 0)
                callback()
              }
            }
            callback()
          },
        })
      } catch (e) {
        tools.log(e)
      } finally {
      }
    },
    // ajax上传文件
    uploadFile: function (url, data, callback) {
      try {
        $.ajax({
          type: 'POST',
          url: url ? url : '',
          data: data ? data : {},
          dataType: 'json',
          contentType: false,
          processData: false,

          success: function (data) {
            if (data.code == 0) {
              callback(data.data)
            } else {
              tools.showMsg(data.msg, 0)
              tools.checkErrCode(data.code)
              callback()
            }
          },
          error: function (error) {
            tools.log(error)
            if (error.readyState != 0) {
              tools.showMsg('网络异常，请稍后重试', 0)
            }
            callback()
          },
        })
      } catch (e) {
        tools.log(e)
      } finally {
      }
    },
    objKeySort: function (obj) {
      var newkey = Object.keys(obj).sort()
      var newObj = {}
      for (var i = 0; i < newkey.length; i++) {
        newObj[newkey[i]] = obj[newkey[i]]
      }
      return newObj
    },
    procRequestParams: function (obj) {
      var keys = Object.keys(obj).sort()
      var bodyArr = []
      for (var k in keys) {
        if (typeof obj[keys[k]] !== 'undefined') {
          var value = decodeURIComponent(obj[keys[k]].toString())
          bodyArr.push(keys[k] + '=' + value)
        } else {
          bodyArr.push(keys[k] + '=')
        }
      }
      return bodyArr.join('&')
    },

    /**
     * 文件上传功能
     * @param { File|Array<File> } files 上传文件或者列表
     * @param { Function } callback  回调地址
     * @param { Object } options 参数配置
     * @param { Object } options--limitWidth  限制宽度
     * @param { Object } options--limitHeight 限制高度
     * @param { Object } options--maxWidth  最大宽度
     * @param { Object } options--maxHeight 最大高度
     * @param { Object } options--maxSize   最大高度
     * @param { RegExp } options--accept    文件类型正则
     * @param { Boolean } options--isFilter    true：检测全部，只有全部正常才会上传；false：上传正常部分
     * @param { String } platform 平台，又拍云(upyun)或群晖(qunhui)；默认又拍云
     * @param { String } path 文件上传文件夹路径
     * 全局配置：
     * NUI.config.tencentCos = {
        bucketName:'', // 桶名
        filePath:'', // 文件路径
        domain:'', //文件访问域名
        isPro: true, //是否正式环境，默认正式环境
        isAZ:true,//是否开启AZ
      }
     */
    // 文件上传
    upload: function (files, callback, options, platform = '', path = '') {
      console.log('NUI.config.tencentCos', NUI.config.tencentCos);
      var bucketName = NUI.config?.tencentCos?.bucketName,
        filePath = NUI.config?.tencentCos?.filePath || '',
        domain = NUI.config?.tencentCos?.domain,
        isPro = !!NUI.config?.tencentCos?.isPro,
        isAZ = !!NUI.config?.tencentCos?.isAZ;
      console.log(NUI.config?.tencentCos?.isPro);
      console.log(isPro);
      if (!bucketName || !domain) {
        tools.showMsg('请配置云存储变量或检查配置是否正确', 0)
        return false;
      }
      kutils.uploadToTencent(files, bucketName, filePath, domain, function (res) {
        if (res.status) {
          callback(res.data, res.files)
        } else {
          callback(false)
        }
      }, isPro, isAZ)
    },
    // 请求错误处理
    checkErrCode: function (code) {
      switch (code) {
        case -1024:
          // 通知后台中心
          var targetOrigin = '*' // 若写成'http://b.com/c/proxy.html'效果一样
          window.parent.postMessage({ code: 1024 }, targetOrigin)
          break
      }
    },

    // 信息展示
    showMsg: function (txt, type) {
      NUI.$toast({
        showClose: true,
        message: txt,
        type: type == 1 ? 'success' : 'warning',
      })
    },
    // 确认框展示
    showConfirm: function (txt, callback, cancel_callback) {
      NUI.$confirm(txt, '提示')
        .then(callback)
        .catch(cancel_callback || function () { })
    },

    // 操作成功页面
    showSuccess: function (url, msg) {
      msg = (msg || '操作成功') + '，正在为您跳转..'
      tools.showMsg(msg, 1)
      setTimeout(function () {
        tools.jump(url, 'keep=1')
      }, 1000)
    },

    // 跳转
    jumpUrl: function (url, params) {
      tools.jump(url, params)
    },

    // 返回列表页并保留搜索条件
    returnIndex: function (url) {
      tools.jump(url, 'keep=1')
    },

    // 跳转
    jump: function (url, params) {
      NUI.postMessage({ type: 'changePage' })
      window.location.href =
        url + (~url.indexOf('?') ? '&' : '?') + (params || '')
    },

    // 新窗口打开连接
    openUrl: function (url) {
      window.open(url)
    },

    // 返回
    jumpBack: function () {
      NUI.postMessage({ type: 'changePage' })
      window.history.go(-1)
    },
    getLineChartsData: function (data) {
      var chartOption = {
        title: {
          text: data.text,
        },
        tooltip: {
          trigger: 'axis',
        },
        legend: {
          data: data.legend,
        },
        grid: {
          left: '3%',
          right: '4%',
          bottom: '3%',
          containLabel: true,
        },
        toolbox: {
          feature: {
            saveAsImage: {},
          },
        },
        xAxis: {
          type: 'category',
          boundaryGap: false,
          data: data.xAxis,
        },
        yAxis: {
          type: 'value',
        },
        series: [],
      }
      for (var i = 0; i < data.series.length; i++) {
        var one = {
          name: data.series[i].name,
          type: 'line',
          stack: data.series[i].stack,
          data: data.series[i].data,
        }
        chartOption.series.push(one)
      }
      return chartOption
    },

    getBarChartsData: function (data) {
      var chartOption = {
        tooltip: {
          trigger: 'axis',
          axisPointer: {
            // 坐标轴指示器，坐标轴触发有效
            type: 'shadow', // 默认为直线，可选为：'line' | 'shadow'
          },
        },
        legend: {
          data: data.legend,
        },
        grid: {
          left: '3%',
          right: '4%',
          bottom: '3%',
          containLabel: true,
        },
        xAxis: [
          {
            type: 'category',
            data: data.xAxis,
          },
        ],
        yAxis: [
          {
            type: 'value',
          },
        ],
        series: [],
      }
      for (var i = 0; i < data.series.length; i++) {
        var one = {
          name: data.series[i].name,
          type: 'bar',
          stack: data.series[i].stack,
          data: data.series[i].data,
        }
        chartOption.series.push(one)
      }
      return chartOption
    },

    /**
     * 复制报表内容到粘贴板
     * @param { Array<Object> } tableData  数据集合
     * @param { Array<String> } headerKey  表头key值
     * @param { Array<String> } headerVal  表头value值
     * @param { Object } options       配置项
     * options
     * @param { String } childrenKey   字段值：children
     * @param { Boolean } useNotation  是否使用科学计数法，默认：true
     */
    copyTableText: function (tableData, headerKey, headerVal, options) {
      var childrenKey
      if ('string' === typeof options) {
        childrenKey = options
        options = arguments[4]
      }
      options = options || {}
      options.childrenKey = childrenKey || options.childrenKey || 'children'

      var text = ''
      headerVal.forEach(function (item) {
        text += item + '\t'
      })
      text += '\n'

      var data = cloneDeep(tableData)
      data.forEach(function (item) {
        text += tools.makeText(headerKey, item, options)
        text += '\n'
        if (item.hasOwnProperty(options.childrenKey)) {
          item[options.childrenKey].forEach(function (children) {
            text += tools.makeText(headerKey, children, options)
            text += '\n'
          })
        }
      })

      tools.copyText(text)
    },

    // 复制文字
    copyText: function (text) {
      // 使用textarea支持换行，使用input不支持换行
      // \t 跳列 \t 换行 在 excel 中适用
      var textarea = document.createElement('textarea')
      textarea.value = text
      document.body.appendChild(textarea)
      textarea.select()
      document.execCommand('copy')
      document.body.removeChild(textarea)

      tools.showMsg('复制成功', 1)
    },

    /**
     * 组装text内容
     * @param { Array<String> } headerKey  表头key值
     * @param { Array<Object> } data       数据集合
     * @param { Object } options       配置项
     * options
     * @param { Boolean } useNotation  是否使用科学计数法，默认：true
     */
    makeText: function (headerKey, data, options) {
      options = options || {}

      var text = ''
      headerKey.forEach(function (key) {
        if (data[key] == undefined) {
          data[key] = ''
        }
        if (options.useNotation !== false) {
          if (
            (!isNaN(data[key]) && data[key].length >= 9) ||
            data[key].toString().indexOf('/') > 0
          ) {
            data[key] = "'" + data[key]
          }
        }
        text += data[key] + '\t'
      })

      return text
    },

    /**
     * 提取格式化参数
     * @param url
     */
    url2Obj: function (url) {
      if (url.indexOf('?') != -1) {
        url = url.split('?')[1]
      }

      var arr = url.split('&')
      var obj = {}
      $.each(arr, function (i, item) {
        var keyVal = item.split('=')
        obj[keyVal[0]] = keyVal[1]
      })

      return obj
    },

    /**
     * 获取cookie
     * @param cookie_name
     * @returns {string}
     */
    getCookie: function (cookie_name) {
      var allcookies = document.cookie
      var cookie_pos = allcookies.indexOf(cookie_name)
      var value = ''
      if (cookie_pos != -1) {
        cookie_pos = cookie_pos + cookie_name.length + 1
        var cookie_end = allcookies.indexOf(';', cookie_pos)
        if (cookie_end == -1) {
          cookie_end = allcookies.length
        }
        value = unescape(allcookies.substring(cookie_pos, cookie_end))
      }

      return value
    },
    setSubmitBtn: function (type) {
      //提交
      if (type == 0) {
        //提交按钮置灰，不可点击
        NUI.data.submitBtn.disabled = true
        NUI.data.submitBtn.text = '正在提交...'
      } else if (type == 1) {
        //取消，提交按钮立即还原
        NUI.data.submitBtn.disabled = false
        NUI.data.submitBtn.text = '提交'
      } else if (type == 2) {
        setTimeout(function () {
          //出错，提交按钮2秒后还原(与消息提示框消息时间同步还原)
          NUI.data.submitBtn.disabled = false
          NUI.data.submitBtn.text = '提交'
        }, 2000)
      }
    },
    /**
     * 月(M)、日(d)、12小时(h)、24小时(H)、分(m)、秒(s)、周(E)、季度(q) 可以用 1-2 个占位符
     * 年(y)可以用 1-4 个占位符，毫秒(S)只能用 1 个占位符(是 1-3 位的数字)
     * java风格
     * @param {Date|Number|String} date 时间
     * @param {String} fmt 默认: yyyy-MM-dd HH:mm:ss"
     * @return {String}
     * @example
     * tools.dateFormat(new Date(), "yyyy-MM-dd hh:mm:ss.S") ==> 2006-07-02 08:09:04.423
     * tools.dateFormat(new Date(), "yyyy-MM-dd E HH:mm:ss") ==> 2009-03-10 二 20:09:04
     * tools.dateFormat(new Date(), "yyyy-MM-dd EE hh:mm:ss") ==> 2009-03-10 周二 08:09:04
     * tools.dateFormat(new Date(), "yyyy-MM-dd EEE hh:mm:ss") ==> 2009-03-10 星期二 08:09:04
     * tools.dateFormat(new Date(), "yyyy-M-d h:m:s.S") ==> 2006-7-2 8:9:4.18
     */
    dateFormat: function (date, fmt) {
      if (!date) return ''
      date =
        date instanceof Date
          ? date
          : isNaN(Number(date))
            ? new Date(date)
            : new Date(Number(date))

      fmt = fmt || 'yyyy-MM-dd HH:mm:ss'
      var o = {
        'M+': date.getMonth() + 1, // 月份
        'd+': date.getDate(), // 日
        'H+': date.getHours(), // 24时制
        'h+': date.getHours() % 12 == 0 ? 12 : date.getHours() % 12, // 12时制
        'm+': date.getMinutes(), // 秒
        's+': date.getSeconds(),
        'q+': Math.floor((date.getMonth() + 3) / 3), // 季度
        S: date.getMilliseconds(), // 毫秒
      }
      var week = {
        0: '/u65e5',
        1: '/u4e00',
        2: '/u4e8c',
        3: '/u4e09',
        4: '/u56db',
        5: '/u4e94',
        6: '/u516d',
      }
      // 年
      if (/(y+)/.test(fmt)) {
        fmt = fmt.replace(
          RegExp.$1,
          (date.getFullYear() + '').substr(4 - RegExp.$1.length)
        )
      }
      // 星期
      if (/(E+)/.test(fmt)) {
        fmt = fmt.replace(
          RegExp.$1,
          (RegExp.$1.length > 1
            ? RegExp.$1.length > 2
              ? '/u661f/u671f'
              : '/u5468'
            : '') + week[date.getDay() + '']
        )
      }
      for (var k in o) {
        if (new RegExp('(' + k + ')').test(fmt)) {
          fmt = fmt.replace(
            RegExp.$1,
            RegExp.$1.length == 1
              ? o[k]
              : ('00' + o[k]).substr(('' + o[k]).length)
          )
        }
      }
      return fmt
    },
    rangeOptions: function () {
      return [
        {
          label: '今日',
          value: function () {
            var now = new Date()
            var start = new Date(
              now.getFullYear(),
              now.getMonth(),
              now.getDate(),
              0,
              0,
              0
            )
            var end = new Date(
              now.getFullYear(),
              now.getMonth(),
              now.getDate(),
              23,
              59,
              59
            )
            return [start, end]
          },
        },
        {
          label: '昨天',
          value: function () {
            var now = new Date()
            var start = new Date(
              now.getFullYear(),
              now.getMonth(),
              now.getDate() - 1,
              0,
              0,
              0
            )
            var end = new Date(
              now.getFullYear(),
              now.getMonth(),
              now.getDate() - 1,
              23,
              59,
              59
            )
            return [start, end]
          },
        },
        {
          label: '7天',
          value: function () {
            var now = new Date()
            var start = new Date(
              now.getFullYear(),
              now.getMonth(),
              now.getDate() - 6,
              0,
              0,
              0
            )
            var end = new Date(
              now.getFullYear(),
              now.getMonth(),
              now.getDate(),
              23,
              59,
              59
            )
            return [start, end]
          },
        },
        {
          label: '14天',
          value: function () {
            var now = new Date()
            var start = new Date(
              now.getFullYear(),
              now.getMonth(),
              now.getDate() - 13,
              0,
              0,
              0
            )
            var end = new Date(
              now.getFullYear(),
              now.getMonth(),
              now.getDate(),
              23,
              59,
              59
            )
            return [start, end]
          },
        },
        {
          label: '30天',
          value: function () {
            var now = new Date()
            var start = new Date(
              now.getFullYear(),
              now.getMonth(),
              now.getDate() - 29,
              0,
              0,
              0
            )
            var end = new Date(
              now.getFullYear(),
              now.getMonth(),
              now.getDate(),
              23,
              59,
              59
            )
            return [start, end]
          },
        },
        {
          label: '上月',
          value: function () {
            var now = new Date()
            var start = new Date(
              now.getFullYear(),
              now.getMonth() - 1,
              1,
              0,
              0,
              0
            )
            var end = new Date(
              new Date(
                now.getFullYear(),
                now.getMonth(),
                1,
                0,
                0,
                0
              ).valueOf() - 1000
            )
            return [start, end]
          },
        },
        {
          label: '本月',
          value: function () {
            var now = new Date()
            var start = new Date(now.getFullYear(), now.getMonth(), 1, 0, 0, 0)
            var end = now
            return [start, end]
          },
        },
      ]
    },
    cloneDeep: function (obj) {
      return cloneDeep(obj);
    },

    /**
     * 判断是否移动端浏览器
     * @returns {boolean}
     */
    isMobile: function () {
      const userAgent = navigator.userAgent.toLowerCase();
      const mobileKeywords = ['mobile', 'android', 'iphone', 'ipod', 'ipad', 'windows phone'];
      return mobileKeywords.some(keyword => userAgent.indexOf(keyword) !== -1);
    },

    arrayMerge: function (arr1, arr2, newKeys = false) {
      let ret = cloneDeep(arr1);
      for (let key in arr2) {
        if (!newKeys && !arr1.hasOwnProperty(key)) {
          continue;
        }
        ret[key] = cloneDeep(arr2[key]);
      }
      return ret;
    },

    // ajax请求
    ajaxGet: function (url, callback, error_msg, timeout = 30000) {
      this.ajax('GET', url, {}, callback, error_msg, timeout, 'application/json');
    },

    // ajax请求
    ajaxPost: function (url, data, callback, error_msg, timeout = 30000) {
      this.ajax('POST', url, JSON.stringify(data), callback, error_msg, timeout, 'application/json');
    },
  }
})(this, document)
