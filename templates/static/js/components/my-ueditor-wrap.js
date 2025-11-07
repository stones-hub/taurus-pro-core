;(function () {
  var scripts = document.getElementsByTagName('script')
  var HOME_URL = scripts[scripts.length - 1].getAttribute('homeUrl')
  var SERVER_URL = scripts[scripts.length - 1].getAttribute('serverUrl')
  // 简单模式默认配置
  var simpleEditor = {
    autoFloatEnabled: false,
    enableAutoSave: false,
    toolbars: [['bold', 'underline', 'strikethrough', 'forecolor', 'cleardoc']],
    elementPathEnabled: false,
    wordCount: false,
    initialFrameHeight: 100,
    initialStyle: 'body{font-size:12px}p{line-height:1.5em}',
    enableContextMenu: true,
    classList: ['simple'],
    zIndex: 1,
  }

  /** 表格多选框复杂组件 */
  NUI.component('my-ueditor-wrap', {
    props: {
      value: {
        type: String,
        default: '',
      },
      /* config 对应 ueditor.config 配置 */
      config: {
        type: Object,
        default: function () {
          return {
            // 是否保持toolbar的位置不动,默认true
            // autoFloatEnabled: true,
          }
        },
      },
      type: {
        type: String,
        default: 'normal',
      },
      disabled: {
        type: Boolean,
        default: false,
      },
    },
    data: function () {
      return {
        id: 'editor_' + Math.random().toString(16).slice(-6),
        isReady: false,
        initValue: this.value,
        defaultConfig: {
          UEDITOR_HOME_URL: /https?:\/\//.test(HOME_URL)
            ? HOME_URL
            : window.UE.getUEBasePath() + HOME_URL,
          serverUrl: SERVER_URL,
          enableAutoSave: false,
          initialFrameWidth: '100%',
          initialFrameHeight: 240,
        },
      }
    },
    computed: {
      mixedConfig: function () {
        var obj = {}
        for (var key in this.defaultConfig) {
          obj[key] = this.defaultConfig[key]
        }
        for (var key in this.config) {
          obj[key] = this.config[key]
        }
        if (this.type === 'simple') {
          if (this.config?.toolbars) {
            const combinedToolbars = [
              ...(simpleEditor.toolbars[0] || []),
              ...(this.config?.toolbars[0] || []),
            ]
            simpleEditor.toolbars = [[...new Set(combinedToolbars)]]
          }
          Object.assign(obj, simpleEditor)
        }
        return obj
      },
    },
    watch: {
      value: function (value) {
        if (this.isReady) {
          if (this.editor.getContent() !== value) {
            this.editor.setContent(value)
          }
        } else {
          this.initValue = value
        }
      },
      disabled: function (value) {
        if (value) {
          this.editor.setDisabled()
        } else {
          this.editor.setEnabled()
        }
      },
    },
    mounted: function () {
      var that = this
      this.editor = window.UE.getEditor(this.id, this.mixedConfig)
      this.editor.addListener('ready', this.readyHandler)

      // 监听 pasteplain 命令是否触发
      this.editor.addListener('beforepaste', function (type, data) {
        if (that.editor.options.pasteplain) {
          let html = data.html
          var div = document.createElement('div')
          div.innerHTML = html
          var blockElements = div.querySelectorAll(
            'h1, h2, h3, h4, h5, h6, div, p, blockquote'
          )
          blockElements.forEach(function (element) {
            var p = document.createElement('p')
            p.innerHTML = element.innerHTML
            p.removeAttribute('style')
            element.replaceWith(p)
          })
          var processedHtml = div.innerHTML
          data.html = processedHtml
          data.text = processedHtml
          data.preventDefault = true
        }
      })

      this.editor.addListener('selectionchange', () => {
        if (this.type !== 'simple') return
        const selectedText = this.editor.selection.getRange()
        if (selectedText.startOffset < selectedText.endOffset) {
          let domUtils = UE.dom.domUtils
          let bk_start = this.editor.selection.getRange().createBookmark().start
          bk_start.style.display = ''
          let x = domUtils.getXY(bk_start).x
          let y = domUtils.getXY(bk_start).y + 20
          $(bk_start).remove()
          // 如果有选中的文本，则显示工具栏
          $(`#${this.id}`)
            .parents('.ui-table-column.is-ellipsis')
            .css({ overflow: 'visible' })
          $(`#${this.id} .edui-editor`).css({
            'white-space': 'nowrap',
            'text-overflow': 'ellipsis',
          })
          $(`#${this.id} .edui-editor-toolbarbox`)
            .removeClass('e-hide')
            .css({
              'box-shadow': 'none',
              'z-index': 11,
              position: 'absolute',
              top: `${y}px`,
              left: `${x}px`,
            })
          $(`#${this.id} .edui-editor-toolbarboxouter`).css({
            'background-image': 'none',
            'border-radius': '5px',
            border: 'none',
            'box-shadow': '0 1px 3px #d4d4d4',
          })
        } else {
          // 如果没有选中的文本，则隐藏工具栏
          $(`#${this.id} .edui-editor-toolbarbox`).addClass('e-hide')
        }
      })
    },
    beforeDestroy: function () {
      if (this.editor) {
        this.editor.removeListener('ready', this.readyHandler)
        this.editor.removeListener('contentChange', this.contentChangeHandler)
        this.editor.destroy()
        this.editor = null
      }
    },
    methods: {
      readyHandler: function () {
        this.disabled && this.editor.setDisabled()
        this.editor.addListener('contentChange', this.contentChangeHandler)
        this.editor.setContent(this.initValue)
        this.isReady = true
        this.$emit('ready', this.editor)
      },
      contentChangeHandler: function () {
        this.$emit('input', this.editor.getContent())
      },
    },
    template:
      '<div class="my-ueditor-wrap"><script :id="id" type="text/plain"></script></div>',
  })
})()
