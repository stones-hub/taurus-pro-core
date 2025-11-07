; (function () {
  /**
   * 表格多选框复杂组件
   * my-checkbox-table 配置说明
   * @param {Array}    value              数据
   * @param {Array}    options            数据集合
   * @param {Boolean}  disabled           是否禁用
   */
  // 添加样式
  NUI.inject('style', {
    innerHTML:
      '\
      .my-checkbox-table { margin-top: -1px; }\
      .my-checkbox-table table { width: 100%; }\
      .my-checkbox-table table tr:hover { background-color: transparent !important; }\
      .my-checkbox-table table td { padding: 4px; }\
      .my-checkbox-table table td:not(:last-child) { text-align: right; width: 140px; }\
      .my-checkbox-table table td:last-child { text-align: left; }\
      .my-checkbox-table table td:last-child .ui-checkbox { min-width: 160px; }\
      .my-checkbox-table table td tr td{padding:0;border:none; border-bottom:1px solid #dadada; border-right:1px solid #dadada;}\
      .my-checkbox-table table td tr td:last-child{border-right:none;}\
      .my-checkbox-table table td tr:last-child td:last-child{border-left:1px solid #dadada;}\
      .my-checkbox-table table td tr:last-child td{border:none;}\
      .my-checkbox-table .ui-checkbox--box-position-right .ui-checkbox__label-text { margin-right: 10px; }\
      .my-checkbox-table .ui-checkbox { display: -webkit-inline-box; display: -ms-inline-flexbox; display: inline-flex; padding: 5px; }\
    ',
  })

  NUI.component('my-checkbox-table', {
    data() {
      return {
        optionsData: null,
      }
    },
    props: {
      disabled: Boolean,
      value: {
        type: Array,
        default: function () {
          return []
        },
      },
      options: {
        type: [Object, Array],
        default: function () {
          return {}
        },
      },
    },
    created() {
      this.filterOptions()
    },
    watch: {
      options() {
        this.filterOptions()
      },
    },
    methods: {
      filterOptions: function () {
        this.optionsData = JSON.parse(JSON.stringify(this.options))
        this.optionsData.map((item) => {
          this.mapTreeData(item)
        })
      },
      mapTreeData: function (item) {
        let children = item.children
        if (children && children.length > 0) {
          children.map((item1) => {
            if (item1.children && item1.children.length > 0) {
              children[0]['isChild'] = true
              this.mapTreeData(item1)
            }
          })
        }
      },
      isSingleCheck: function (child) {
        var id = child.id
        return -1 !== this.value.indexOf(id)
      },
      isAllCheck: function (option) {
        if (option instanceof Array) {
          for (var i = 0; i < option.length; i++) {
            if (!this.isAllCheck(option[i])) return false
            this.onSingleClick(option[i].id, option[i], true)
          }
          return true
        } else if (option.children) {
          return this.isAllCheck(option.children)
        } else {
          return ~this.value.indexOf(option.id)
        }
      },
      onSingleClick: function (val, child, isAll) {
        var id = child.id
        var index = this.value.indexOf(id)
        if (val) {
          if (-1 == index) {
            this.value.push(id)
          }
        } else {
          if (-1 != index) {
            this.value.splice(index, 1)
          }
        }
        if (!isAll) {
          this.$emit('input', this.value)
        }
      },
      onAllClick: function (val, option, isRecursive) {
        if (option instanceof Array) {
          for (var i = 0; i < option.length; i++) {
            this.onSingleClick(val, option[i], true)
            this.onAllClick(val, option[i], true)
          }
        } else if (option.children) {
          this.onSingleClick(val, option, true)
          this.onAllClick(val, option.children, true)
        } else {
          this.onSingleClick(val, option, true)
        }
        if (!isRecursive) {
          this.$emit('input', this.value)
        }
      },
    },
    template:
      '<div class="my-checkbox-table ui-table-layout">\
          <table>\
            <tbody>\
              <template v-for="(option, index) in optionsData">\
                <template v-if="option.children && option.children.length>0 && option.children[0].isChild">\
                  <tr v-for="(option2, index2) in option.children" :key="option.name + option2.name">\
                    <td v-if="0==index2" :rowspan="option.children.length"><ui-checkbox color="green" :value="isAllCheck(option)" position="right" @change="function(evt){return onAllClick.call(this, evt, option);}" :disabled="disabled">{{option.name}}</ui-checkbox></td>\
                    <td><ui-checkbox color="green" :value="option2.children?isAllCheck(option2):isSingleCheck(option2)" position="right" @change="function(evt){return onAllClick.call(this, evt, option2);}" :disabled="disabled">{{option2.name}}</ui-checkbox></td>\
                    <template v-if="option2.children && option2.children.length>0 && option2.children[0].isChild">\
                      <td colspan="2" style="padding:0"><table><tr v-for="option3 in option2.children" :key="option3.id">\
                        <td :colspan="option3.children ? 1 : 2"><ui-checkbox color="green" :value="option3.children?isAllCheck(option3):isSingleCheck(option3)" position="right" @change="function(evt){return onAllClick.call(this, evt, option3);}" :disabled="disabled">{{option3.name}}</ui-checkbox></td>\
                        <td v-if="option3.children"><div class="ui-grid"><ui-checkbox color="green" v-for="option4 in option3.children" :key="option4.id" :value="isSingleCheck(option4)" @change="function(evt){return onSingleClick.call(this, evt, option4);}" :disabled="disabled">{{option4.name}}</ui-checkbox></div></td>\
                      </tr></table></td>\
                    </template>\
                    <td v-else colspan="2"><div class="ui-grid"><ui-checkbox color="green" v-for="(child, index3) in option2.children" :key="child.id" :value="isSingleCheck(child)" @change="function(evt){return onSingleClick.call(this, evt, child);}" :disabled="disabled">{{child.name}}</ui-checkbox></div></td>\
                  </tr>\
                </template>\
                <tr :key="option.id" v-else>\
                  <td><ui-checkbox color="green" :value="isAllCheck(option)" position="right" @change="function(evt){return onAllClick.call(this, evt, option);}" :disabled="disabled">{{option.name}}</ui-checkbox></td>\
                  <td colspan="3"><div class="ui-grid"><ui-checkbox color="green" v-for="(child, index3) in option.children" :key="child.id" :value="isSingleCheck(child)" @change="function(evt){return onSingleClick.call(this, evt, child);}" :disabled="disabled">{{child.name}}</ui-checkbox></div></td>\
                </tr>\
              </template>\
              <tr v-if="!disabled">\
                <td><ui-checkbox color="green" :value="isAllCheck(options)" position="right" @change="function(evt){return onAllClick.call(this, evt, options);}">全选</ui-checkbox></td>\
                <td colspan="3"></td>\
              </tr>\
            </tbody>\
          </table>\
        </div>',
  })
})()
