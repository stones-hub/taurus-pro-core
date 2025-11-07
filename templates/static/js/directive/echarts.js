(function() {
    function deepObjectMerge(firstObj, secondObj) { // 深度合并对象
        for (var key in secondObj) {
            firstObj[key] = firstObj[key] && firstObj[key].toString() === "[object Object]" ?
                deepObjectMerge(firstObj[key], secondObj[key]) : firstObj[key] = secondObj[key];
        }
        return firstObj;
    }

    var axisLineColor = '#909399'
    function getOptionWithDefault(option) {
        option = option || {}
        var baseOption = {
            // color: ['#55acee', '#4cdfc0', '#ff9945', '#facf2a', '#788cf0', '#2dca93', '#ffc57f', '#bda29a', '#6e7074', '#546570', '#c4ccd3'],
            color: option.colors || [
                '#55acee',
                '#6be6c1',
                '#626c91',
                '#a0a7e6',
                '#c4ebad',
                '#96dee8',
                '#ffc57f',
                '#bda29a',
                '#f5e8c8',
                '#546570',
                '#c4ccd3'
            ],
            grid: {
                top: '30',
                left: '60',
                right: '60',
                bottom: '60'
            },
            tooltip: {
                trigger: 'axis',
                axisPointer: {
                    type: 'cross'
                }
            },
            legend: {
                data: [],
                bottom: '0'
            },
            xAxis: {
                type: 'category',
                axisLine: {
                    lineStyle: { color: option.xAxisNone ? 'transparent': axisLineColor  }
                },
                data: []
                    // axisPointer: {
                    //   type: 'shadow'
                    // }
            },
            yAxis: [
                {
                    type: 'value',
                    max: null,
                    axisLine: {
                        lineStyle: { color:option.yAxisNone ? 'transparent': axisLineColor }
                    }
                },
                {
                    type: 'value',
                    max: null,
                    axisLine: {
                        lineStyle: { color:option.yAxisNone ? 'transparent': axisLineColor }
                    }
                }
            ],
            series: []
        }
        return deepObjectMerge(baseOption, option)
    }

    function appendMask(el) {
        removeMask(el)
        var mask = document.createElement('div')
        mask.className = 'echarts-mask'
        mask.innerHTML = '<span style="margin-top: -50px;color: #909399">暂无数据</span>'
        mask.style = 'position: absolute; left: 0; top: 0; width: 100%; height: 100%;display: flex; align-items: center; justify-content: center;'
    
        el.appendChild(mask)
    }
    
    function removeMask(el) {
        var mask = el.querySelector('.echarts-mask')
        mask && el.removeChild(mask)
    }

    // 注册v-echarts指令
    NUI.directive('echarts', {
        bind: function(el) {
            el.resizeEventHandler = function() {
                return el.echartsInstance.resize()
            }
    
            if (window.attachEvent) {
                window.attachEvent('onresize', el.resizeEventHandler)
            } else {
                window.addEventListener('resize', el.resizeEventHandler, false)
            }
        },
    
        inserted: function(el, binding) {
            el.echartsInstance = echarts.init(el)
            el.__echartsOption = binding.value;
            el.echartsInstance.setOption(getOptionWithDefault(el.__echartsOption));
        },

        update: function(el, binding) {
            try {
                if (!binding.value.series ||
                    !binding.value.series.length ||
                    !binding.value.series[0].data ||
                    !binding.value.series[0].data.length
                ) {
                    appendMask(el)
                } else {
                    removeMask(el)
                }
            } catch (err) {
                console.log(err)
            }
            if(el.__echartsOption !== binding.value) {
                // 清除功能
                if(el.getAttribute('clear')) {
                    el.echartsInstance.dispose()
                    el.echartsInstance = echarts.init(el)
                }
                el.__echartsOption = binding.value;
                el.echartsInstance.setOption(getOptionWithDefault(el.__echartsOption))
            }
        },
    
        unbind: function(el) {
            if (window.attachEvent) {
                window.detachEvent('onresize', this.resizeEventHandler)
            } else {
                window.removeEventListener('resize', this.resizeEventHandler, false)
            }
    
            el.echartsInstance.dispose()
        }
    })
})();