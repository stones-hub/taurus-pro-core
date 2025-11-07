;(function () {
  /**
   * my-simple-table 配置说明
   * @param {Boolean} loading           是否显示加载进度条，默认false
   * @param {Array}   columns           列配置，详细见下表
   * @param {Array}   data              数据集合
   * @param {Array}   selects           复选框值，当使用type=checkbox，用来保存选中的data，name指定data的唯一key值
   * @param {Array}   sorts             排序状态值，当存在排序功能时，用来保存排序的状态，name指定data的唯一key值
   * @param {Boolean} useInnerModel     是否使用内部的model，默认false
   * @param {String}  height            表格的整体高度，默认80vh，既屏幕的80%
   * @param {String}  stickyContainer   吸顶滚动容器，默认 #app
   * @param {Number}  stickyTop         吸顶距离，默认 0
   */
  /**
   * columns 配置说明
   * @param {String}  title      列名
   * @param {String}  fixed      列固定：left-向左，right-向右，true-同left
   * @param {String}  slot       插槽，当slot存在时，name值无效
   * @param {String}  name       列名key值
   * @param {String}  className  样式名（列头：[className]-th，列内容：[className]-td）
   * @param {String}  align      文本方向：left-向左，center-居中，right-向右
   * @param {String}  type       显示类型，默认文本，checkbox-复选框 img-图片，video-视频，switch-开关，link-页面跳转，drawer-抽屉页面，-抽屉页面
   * @param {String}  width      自定义宽度
   * @param {Boolean} ellipsis   超出文本是否省略，需要和width配合使用，默认：true
   * @param {String}  poster     当type为video，该值为视频默认图片的key值
   * @param {String}  href       当type为link，该值为页面跳转路径的key值
   * @param {String}  src        当type为drawer，该值为抽屉打开页面路径的key值
   * @param {Boolean} visible    显隐：true or false
   */
  // 添加样式
  NUI.inject('style', {
    innerHTML:
      '\
    .my-table-drawer .ui-drawer__header { position: absolute; top: 0; right: 0; left: 0; }\
    .my-table-modal .ui-modal__body { position: relative; padding: 0 !important; font-size: 0; }\
    .my-table-drawer .ui-progress-linear, .my-table-modal .ui-progress-linear { position: absolute; top: 0; left: 0; right: 0; height: 2px; }\
    .my-table-drawer .ui-progress-linear__progress-bar, .my-table-modal .ui-progress-linear__progress-bar { height: 2px; }\
  ',
  })

  /** 表格列表组件 */
  NUI.component('my-simple-table', {
    name: 'my-simple-table',
    props: {
      loading: Boolean,
      columns: {
        type: Array,
        default: function () {
          return []
        },
      },
      data: {
        type: Array,
        default: function () {
          return []
        },
      },
      footerData: {
        type: [Function, Object, Array],
        default: null,
      },
      /* 复选框 */
      selects: {
        type: Array,
        default: function () {
          return []
        },
      },
      /* 排序 */
      sorts: {
        type: Object,
        default: function () {
          return {}
        },
      },
      filterMethod: {
        type: Object,
        default: function () {
          return {}
        },
      },
      hasSearch: {
        type: Boolean,
        default: false,
      },
      customHeaderRow: {
        type: [Function, Object],
        default: null,
      },
      customHeaderCell: {
        type: [Function, Object],
        default: null,
      },
      customRow: {
        type: [Function, Object],
        default: null,
      },
      customCell: {
        type: [Function, Object],
        default: null,
      },
      customFooterRow: {
        type: [Function, Object],
        default: null,
      },
      customFooterCell: {
        type: [Function, Object],
        default: null,
      },
      useInnerModel: Boolean,
      /* 表格的高度 */
      height: {
        type: String,
        default: '80vh',
      },
      /* 吸顶距离 */
      stickyContainer: {
        type: String,
        default: '#app', // 使用时根据滚动位置设置，例如#app
      },
      stickyTop: {
        type: Number,
        default: 0,
      },
    },
    computed: {
      tfColumns: function () {
        var self = this
        var tfColumn = function (column) {
          var obj = expansion({}, column, {
            dataIndex: column.dataIndex || column.name,
          })

          if (column.type === 'checkbox') obj.type = 'selection'
          if (column.slot)
            obj.renderCell = function (props) {
              var render = self.$scopedSlots[column.slot]
              return render
                ? render(
                    expansion({}, props, {
                      $index: props.rowIndex,
                      data: props.row,
                    })
                  )
                : null
            }
          else if (column.type) obj.renderCell = column.type

          if (column.renderHeader)
            obj.renderHeader =
              self.$scopedSlots[column.renderHeader] || obj.renderHeader
          if (column.renderCell)
            obj.renderCell =
              self.$scopedSlots[column.renderCell] || obj.renderCell
          if (column.renderFooter)
            obj.renderFooter =
              self.$scopedSlots[column.renderFooter] || obj.renderFooter
          if (column.children) obj.children = column.children.map(tfColumn)

          return obj
        }

        return this.columns.map(tfColumn)
      },
    },
    data: function () {
      return {
        /* 抽屉数据 */
        drawer: {
          loading: false, // 加载状态
          visible: false, // 显示状态
          src: '',
        },
        /* 弹窗尺寸 */
        modal: {
          loading: false, // 加载状态
          size: 'auto', // 弹窗大小
          src: '', //
          title: '标题', //
          width: '', //
          height: '', //
        },
      }
    },
    methods: {
      /* 展示媒体资源 */
      showMedia: function (type, src, title) {
        this.$message(
          {
            mode: '',
            title: title || '',
            size: 'auto',
            content: function (h) {
              return h(type, {
                ref: 'video' === type ? 'media' : '',
                attrs: {
                  style:
                    'max-height: ' +
                    ('video' === type ? '80vh' : '') +
                    '; max-width: 80vw; object-fit: contain;',
                  controls: true,
                  src: src,
                },
              })
            },
          },
          {
            mounted: function () {
              if (this.$refs.media && this.$refs.media.paused) {
                var fn = this.$refs.media.play()
                fn && fn.catch(() => {})
              }
            },
          }
        ).catch(function () {})
      },
      /* 打开抽屉 */
      openDrawer: function (src, column) {
        var self = this
        self.drawer.visible = true
        setTimeout(function () {
          self.drawer.loading = true
          self.drawer.src = src
        }, 100)
      },
      /** 打开弹窗 */
      openModal: function (src, column) {
        var self = this
        var modal = column.modal || {}

        /** 是否使用内部弹窗 */
        if (self.useInnerModel || window.top === window.parent) {
          self.modal.loading = true
          self.modal.title = modal.title || '标题'
          self.modal.width = modal.width
          self.modal.height = modal.height
          self.modal.size = modal.size || 'auto'
          setTimeout(function () {
            self.modal.src = src
          }, 100)
          self.$refs['modal'].open()
        } else {
          NUI.postMessage({
            type: 'tableModal',
            data: {
              title: modal.title || '标题',
              width: modal.width,
              height: modal.height,
              size: modal.size || 'auto',
              src: src,
            },
          })
        }
      },
      getCustomHeaderCell: function (record) {
        var column = record.column
        var props = wrapFn(this.customHeaderCell)(record) || {}
        if (column.className) {
          props.class = [].concat(props.class, {
            [column.className + 'th']: true,
          })
        }
        return props
      },
      getCustomCell: function (record) {
        var row = record.row,
          column = record.column
        var props = wrapFn(this.customCell)(record) || {}
        // 带背景颜色
        if (column.type === 'color') {
          props.style = expansion({}, props.style, {
            backgroundColor:
              'rgba(76,175,80,' + parseFloat(row[column.name]) / 100 + ')',
          })
        }
        if (column.className) {
          props.class = [].concat(props.class, {
            [column.className + 'td']: true,
          })
        }
        return props
      },
      onUpdateSelects: function (selects) {
        this.$emit('update:selects', selects)
      },
      onUpdateSorts: function (order, column) {
        this.$emit('sort', order, column)
      },
      onUpdateFilter: function (value, column) {
        this.$emit('filter-method', value, column)
      },
    },
    template:
      '<div class="my-simple-table">\
        <ui-table v-bind="$props" class="ui-simple-table" :columns="tfColumns" :custom-header-cell="getCustomHeaderCell" :custom-cell="getCustomCell" :hasSearch="hasSearch" :sticky-offset="{top: stickyTop, bottom: 0}" @select="onUpdateSelects" @sort="onUpdateSorts" @filter-method="onUpdateFilter">\
          <!-- 图片显示 -->\
          <my-media slot="img" slot-scope="scope" type="img" :src="scope.row[scope.column.name]" @click="showMedia(\'img\', scope.row[scope.column.name])"></my-media>\
          <!-- 视频显示 -->\
          <my-media slot="video" slot-scope="scope" type="video" :src="scope.row[scope.column.name]" :poster="scope.row[scope.column.poster]" @click="showMedia(\'video\', scope.row[scope.column.name])"></my-media>\
          <!-- 图片/视频显示，不确定素材类型 -->\
          <my-media slot="resource" slot-scope="scope" :type="scope.row[scope.column.name].type" :src="scope.row[scope.column.name].src" :poster="scope.row[scope.column.name].poster" @click="showMedia(`${scope.row[scope.column.name].type}`, scope.row[scope.column.name].src)"></my-media>\
          <!-- 开关控件 -->\
          <ui-switch slot="switch" slot-scope="scope" v-model="scope.row[scope.column.name]" color="green" :true-value="1" :false-value="0" @change="$emit(\'switch\', $event, scope.row, scope.column)"></ui-switch>\
          <!-- 点击跳转 -->\
          <my-link slot="link" slot-scope="scope" @click="$emit(\'jump\', scope.row[scope.column.href])" :text="scope.row[scope.column.name]"></my-link>\
          <!-- 打开抽屉 -->\
          <my-link slot="drawer" slot-scope="scope" @click="openDrawer(scope.row[scope.column.src], scope.column)" :text="scope.row[scope.column.name]"></my-link>\
          <!-- 打开弹窗 -->\
          <my-link slot="modal" slot-scope="scope" @click="openModal(scope.row[scope.column.src], scope.column)" :text="scope.row[scope.column.name]"></my-link>\
        </ui-table>\
        <!-- 抽屉页面 -->\
        <ui-drawer class="my-table-drawer" size="90%" :visible.sync="drawer.visible" direction="rtl" @close="drawer.src=\'\'">\
            <!-- 加载进度条 -->\
            <ui-progress-linear v-show="drawer.loading" color="primary" type="indeterminate"></ui-progress-linear>\
            <iframe v-if="drawer.visible && drawer.src" :src="drawer.src" width="100%" height="100%" frameborder="0" @load="drawer.loading=false" @error="drawer.loading=false"></iframe>\
        </ui-drawer>\
        <!-- 弹窗页面 -->\
        <ui-modal class="my-table-modal" ref="modal" :size="modal.size" :title="modal.title" @close="modal.src=\'\'">\
            <!-- 加载进度条 -->\
            <ui-progress-linear v-show="modal.loading" color="primary" type="indeterminate"></ui-progress-linear>\
            <iframe v-if="modal.src" :style="{width:modal.width, height:modal.height}" :src="modal.src" width="100%" height="100%" frameborder="0" @load="modal.loading=false" @error="modal.loading=false"></iframe>\
        </ui-modal>\
    </div>',
    components: {
      'my-media': {
        props: {
          type: String,
          src: String,
          poster: String,
        },
        render(h) {
          var self = this
          return this.src
            ? h('ui-media', {
                attrs: {
                  type: self.type,
                  src: self.src,
                  poster: self.poster,
                  width: 40,
                  height: 25,
                  // popoverWidth: 300,
                  hoverEffect: 'popover',
                  showEffect: 'video' === self.type ? 'cover' : 'contain',
                },
                on: {
                  click: function (evt) {
                    self.$emit('click', evt)
                  },
                },
              })
            : h('span', {
                domProps: { innerHTML: '-' },
              })
        },
      },
      'my-link': {
        props: {
          text: String,
        },
        render(h) {
          var self = this
          return this.text
            ? h('a', {
                attrs: {
                  href: 'javascript:;',
                },
                on: {
                  click: function (evt) {
                    self.$emit('click', evt)
                  },
                },
                domProps: { innerHTML: this.text },
              })
            : h('span', {
                domProps: { innerHTML: '-' },
              })
        },
      },
    },
  })

  function expansion(target) {
    var objs = Array.prototype.slice.call(arguments, 1)
    for (var i = 0; i < objs.length; i++) {
      var source = objs[i]
      for (var key in source) {
        if (source.hasOwnProperty(key)) {
          target[key] = source[key]
        }
      }
    }
    return target
  }

  function wrapFn(fn) {
    if ('function' === typeof fn) {
      return fn
    }
    return function () {
      return fn
    }
  }
})()
