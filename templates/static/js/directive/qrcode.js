(function() {
    function deepObjectMerge(firstObj, secondObj) { // 深度合并对象
        for (var key in secondObj) {
            firstObj[key] = firstObj[key] && firstObj[key].toString() === "[object Object]" ?
                deepObjectMerge(firstObj[key], secondObj[key]) : firstObj[key] = secondObj[key];
        }
        return firstObj;
    }

    // 注册v-qrcode指令
    NUI.directive('qrcode', {
        inserted: function(el, binding) {
            el.qrcodeInstance = new QRCode(el, binding.value)
        },

        update: function(el, binding) {
            if (typeof binding.value == 'string') {
                el.qrcodeInstance.makeCode(binding.value)
            } else {
                el.qrcodeInstance._htOption = deepObjectMerge(el.qrcodeInstance._htOption, binding.value);
                el.qrcodeInstance.makeCode(binding.value.text || '');
            }
        },
    
        unbind: function(el) {
            el.qrcodeInstance = null;
        }
    })
})();