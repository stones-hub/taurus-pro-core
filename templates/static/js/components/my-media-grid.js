;(function () {
  /**
   * my-media-grid 配置说明
   * @param {Object}   keys              字段配置，详细见下表
   * @param {String}   type              类型 img / video
   * @param {Array}    data              数据集合
   * @param {Number}   col               格子数量
   * @param {Number}   height            格子高度，默认160
   * @param {Number}   backgroundColor   背景颜色，默认#dfdfdf
   */
  /**
   * keys 配置说明
   * @param {String}  title      表头名
   * @param {String}  type       类型 img / video
   * @param {String}  src        路径名
   * @param {String}  poster     视频预览图
   */
  // 添加样式
  NUI.inject('style', {
    innerHTML:
      '\
    .my-media-grid ul { width: 100%; font-size: 0; }\
    .my-media-grid li { display: inline-block; margin: 0 .5% 1%; font-size: 12px; background:#FFF;border:1px solid;border-color: rgb(245, 247, 250); padding: 0px; border-radius: 10px; -webkit-transition: all 400ms; transition: all 400ms; }\
    .my-media-grid li:hover {background:#F8F8F8;-webkit-box-shadow: 0 4px 10px rgba(0, 0, 0, 0.15); box-shadow: 0 4px 10px rgba(0, 0, 0, 0.15); }\
    .my-media-grid li .ui-button { height: 24px; }\
    .ui-media{background:red;}\
    .media-slot-time{position: absolute; bottom: 5px; right: 5px; background: rgba(0,0,0,.5); color: #fff; padding: 2px 5px; border-radius: 5px; }\
    .my-media-grid li:hover .media-slot-time{display: none;}\
  ',
  })

  /** 表格列表组件 */
  NUI.component('my-media-grid', {
    name: 'my-media-grid',
    props: {
      loading: Boolean,
      keys: {
        type: Object,
        default: function () {
          return {
            type: 'type',
            src: 'src',
            poster: 'poster',
            title: 'title',
          }
        },
      },
      data: {
        type: Array,
        default: function () {
          return []
        },
      },
      type: {
        type: String,
        default: 'img',
      },
      col: {
        type: Number,
        default: 6,
      },
      /* 格子的高度 */
      height: {
        type: Number,
        default: 160,
      },
      backgroundColor: {
        type: String,
        default: '#DFDFDF',
      },
    },
    mounted: function () {
      console.log(this.data)
    },
    methods: {
      /* 展示媒体资源 */
      showMedia: function (type, src, audio = '') {
        this.$message({
          mode: '',
          title: '',
          size: 'auto',
          content: function (h) {
            return h(type == 'carousel' ? 'ui-media-carousel' : type, {
              attrs: {
                style:
                  'max-height: ' +
                  ('video' === type ? '100vh' : '') +
                  '; max-width: 100vw; object-fit: contain;',
                controls: true,
                src: src,
                audio: audio,
              },
            })
          },
        })
      },
      hoverEffect: function (type) {
        return type === 'video' ? 'play' : 'popover'
      },
    },
    template:
      '<div class="my-media-grid">\
      <ul>\
          <li v-for="(item, index) in data" :key="index" :style="{width: (100/col-1)+\'%\'}">\
              <slot :index="index" :data="item">\
                  <div style="position:relative;">\
                  <div style="position:relative;">\
                    <ui-media width="100%" :height="height" :background-color="backgroundColor" :hover-effect="hoverEffect(item[keys.type]||type)" :src="item[keys.src]" :type="item[keys.type] || type" :poster="item[keys.poster]" :audio="item[keys.audio]" @click="showMedia(item[keys.type] || type, item[keys.src], item[keys.audio])" style="padding:5px 0;border-radius: 10px 10px 0 0;overflow:hidden;"></ui-media>\
                    <div v-if="$slots.time || $scopedSlots.time" class="media-slot-time">\
                      <slot name="time" :index="index" :data="item"></slot>\
                    </div>\
                    </div>\
                    <slot name="tag" :index="index" :data="item"></slot>\
                  </div>\
                  <slot name="info" :index="index" :data="item"></slot>\
              </slot>\
          </li>\
      </ul>\
    </div>',
  })
})()
