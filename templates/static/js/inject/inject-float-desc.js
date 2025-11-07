;(function(){
    /* 自定义注入组件 */
    
    // 添加样式
    NUI.inject('style', { innerHTML: '\
        .inject-float-desc { position:absolute; z-index: 10; right:30px; top:50%; width:80px; height:86px; background-image: url(data:image/png;base64,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);  background-repeat: no-repeat; background-size: 100%; }\
        .inject-float-desc .link { display: block; margin-top: 64px; width: 100%;  height: 22px;  text-align: center; font-size: 12px; color: #fff; }\
        .inject-float-desc .link:hover { text-decoration: underline; }\
        .inject-float-desc .clear { position: absolute; top: -2px; right: -2px; padding: 2px; width: 20px; height: 20px; font-size: 16px; border-radius: 50%; background-color: rgba(0,0,0,.5); color:#fff; cursor: pointer;}\
    '});

    /**
     * 添加组件逻辑，由于NUI是单例模式，所以组件模式内部的数据都必须使用this，或者self托管来调用方法以及数据
     * 禁止使用NUI.data 和 NUI.__app__，即使使用了，也是调用外框的属性方法
     */ 
    NUI.component('inject-float-desc', {
        data: function() {
            return {
                url: './page/version.html',
                text: '版本更新',

                isShow: true,
                isClear: false,
            }
        },
        methods: {
            onClick() {
                // this.$root可以读取到上层数据，或者直接使用NUI.data 和 NUI.__app__
                this.$root.changePage(this.url);
            }
        },
        template: '<div v-if="isShow" class="inject-float-desc" @mouseenter="isClear=true" @mouseleave="isClear=false" v-drag="\'body\'">\
            <a class="link" @mousedown.stop @click="onClick">{{text}}</a>\
            <ui-icon v-show="isClear" class="clear" icon="clear" @click="isShow=false"></ui-icon>\
        </div>'
    })

    // 将组件名注入到框架中，这一句一定要写，否则不会生效。
    NUI.data.injectComponents.push('inject-float-desc');
})();