let url = ''
if (/8088/.test(location.origin)) {
  url = '/fonts/spark-md5.min.js'
} else if (/web-test/.test(location.origin)) {
  url = '/nui/lib/fonts/spark-md5.min.js'
} else {
  url = '/static/nui/lib/fonts/spark-md5.min.js'
}
self.importScripts(url) // 导入脚本
// 生成文件 hash
self.onmessage = (e) => {
  const { fileChunkList } = e.data
  const spark = new self.SparkMD5.ArrayBuffer()
  let percentage = 0
  let count = 0
  const loadNext = (index) => {
    const reader = new FileReader()
    reader.readAsArrayBuffer(fileChunkList[index].file)
    reader.onload = (e) => {
      count++
      spark.append(e.target.result)
      if (count === fileChunkList.length) {
        self.postMessage({
          percentage: 100,
          hash: spark.end(),
        })
        self.close()
      } else {
        percentage += 100 / fileChunkList.length
        self.postMessage({
          percentage,
        })
        loadNext(count)
      }
    }
  }
  loadNext(0)
}
