;(function(){
    /* 我的下载弹窗页面 */

    // 添加样式
    NUI.inject('style', { innerHTML: '\
        .inject-my-download .wrap { width: 900px; }\
    '});

    /**
     * 添加组件逻辑，由于NUI是单例模式，所以组件模式内部的数据都必须使用this，或者self托管来调用方法以及数据
     * 禁止使用this 和 this，即使使用了，也是调用外框的属性方法
     */
    NUI.component('inject-my-download', {
        data: function() {
            return {
                url: {
                    myDownload: '/admin/index/myDownload',
                    download: '/admin/index/download',
					rerun: '/admin/index/rerun',
                    getUnreadMsg: '/admin/index/getUnreadMsg'
                },
                form: {
                    date: [],
                    page_no: 1,
                    page_size: 15,
                },
                columns: [
                    { title: '序号', name: 'id', align: 'center', visible: true },
                    { title: '文件名称', name: 'filename', align: 'center', visible: true },
                    { title: '文件大小', name: 'file_size', align: 'center', visible: true },
                    { title: '操作时间', name: 'created_at', align: 'center', visible: true },
                    { title: '状态', name: 'status_name', align: 'center', visible: true },
                    { title: '操作', slot: 'options', align: 'center', visible: true },
                ],
                table: {
                    data: [],
                    loading: false,
                    result_total: 0,
                    pageSizes: [15, 30, 50, 100] // 分页配置
                },

                /* 时间控件-快捷方式 */
                rangeOptions: tools.rangeOptions(),
            }
        },
        created: function() {
            NUI.__app__.$on('open:avatar-dropdown', this.getUnreadMsg);
        },
        beforeCreate: function() {
            NUI.__app__.$off('open:avatar-dropdown', this.getUnreadMsg);
        },
        methods: {
            getUnreadMsg: function () {
                var getRandomInt = function(n) { return Math.round(Math.random() * n) }
                ///////////// 随机数据 /////////////
                $.each(NUI.data.avatarMenu, function(index, item) {
                    if(item.name === 'download') {
                        item.badge = getRandomInt(100);
                        return false;
                    }
                })
            },
            searchMyDownloadData: function () {
                this.form.page_no = 1
                this.loadMyDownloadData()
            },
            loadMyDownloadData: function() {
                var self = this;
                self.table.loading = true;
                // 模拟ajax异步
                setTimeout(function () {
                  self.table.loading = false;
                  // 总条数
                  self.table.result_total = 4
                  self.table.data = $.map(new Array(+self.form.page_size), function() {
                    return {
                      id: 27,
                      filename: '日留存.csv',
                      file_size: '27.4K',
                      created_at: '2021-01-05 16:36:44',
                      status_name: '已完成',
                      status: 2,
                    }
                  })
                }, 1200)
                this.getUnreadMsg()
            },
            download: function (id) {
                var url = this.url.download  + '?id=' + id
                tools.openUrl(url, 0);
            },
            rerun: function (id) {
                tools.showMsg('操作成功', 1)
                this.loadMyDownloadData()
            },
            pageSizeChange: function (val) {
                this.form.page_size = Number(val);
                this.loadMyDownloadData();
            },
            pageNumChange: function (val) {
                this.form.page_no = val;
                this.loadMyDownloadData();
            },
            getCustomRow: function ({row, rowIndex}) {
                if(row.is_read == 0) {
                    return {
                        style: {
                            fontWeight: 'bold'
                        }
                    }
                }
            },
            getCustomCell: function ({row, column, rowIndex, columnIndex}) {
                if (column.name == 'status_name') {
                    return {
                        style: {
                            color: row.status == 2 ? 'green' : row.status == 3 ?  'red' : ''
                        }
                    }
                }
            },
        },
        template: '<ui-modal class="inject-my-download" ref="myDownload" size="auto" title="我的下载" @open="loadMyDownloadData">\
            <div class="wrap">\
                <!-- 条件筛选栏 -->\
                <section class="my-search">\
                    <div class="ui-grid">\
                        <div class="ui-grid-cell col-6">\
                            <ui-date-time-picker placeholder="操作时间" type="daterange" :options="rangeOptions" v-model="form.date" format="yyyy-MM-dd"></ui-date-time-picker>\
                        </div>\
                        <div class="ui-grid-cell col-4">\
                            <ui-button color="primary" icon="search" @click="searchMyDownloadData">查询</ui-button>\
                        </div>\
                        <div class="ui-grid-cell ui-grid-full ui-grid-middle ui-text-right">\
                            <ui-button color="primary" icon="refresh" @click="loadMyDownloadData">刷新数据</ui-button>\
                        </div>\
                    </div>\
                </section>\
                <section class="my-table">\
                    <my-simple-table :columns="columns" :data="table.data" :loading="table.loading" :custom-row="getCustomRow" :custom-cell="getCustomCell">\
                        <template slot="options" slot-scope="scope">\
                            <ui-button v-if="scope.data.status == 2" type="text" color="primary" @click="download(scope.data.id)">下载</ui-button>\
                            <ui-button v-if="scope.data.status == 0 || scope.data.status == 3" type="text" color="primary" @click="rerun(scope.data.id)">重新执行</ui-button>\
                        </template>\
                    </my-simple-table>\
                </section>\
                <section class="my-pagination">\
                    <ui-pagination :page-size="form.page_size" :total="table.result_total" :current="form.page_no" @change="pageNumChange"></ui-pagination>\
                    <span>总{{Math.ceil(table.result_total/form.page_size)}}页/{{table.result_total}}条记录 每页显示</span>\
                    <ui-select v-model="form.page_size" :options="table.pageSizes" @change="pageSizeChange"></ui-select>\
                </section>\
            </div>\
        </ui-modal>'
    })

    // 将组件名注入到框架中，这一句一定要写，否则不会生效。
    NUI.data.injectComponents.push('inject-my-download');
})();
