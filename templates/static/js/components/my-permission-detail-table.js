; (function () {
  /**
   * 权限详情表格组件（只读）
   * my-permission-detail-table 配置说明
   * @param {Array}  permissions  权限树数据，支持层级结构
   * 
   * 使用示例：
   * <my-permission-detail-table 
   *   :permissions="permissions"
   * ></my-permission-detail-table>
   * 
   * 权限数据结构：
   * {
   *   permission_id: 1,
   *   permission_name: "系统管理",
   *   permission_code: "system",
   *   permission_type: 1,
   *   checked: true,
   *   children: []
   * }
   */
  
  // 添加样式
  NUI.inject('style', {
    innerHTML:
      '\
      .my-permission-detail-table { width: 100%; border-collapse: collapse; font-size: 13px; border: 1px solid #d9d9d9; table-layout: auto; }\
      .my-permission-detail-table th { background-color: #fafafa; padding: 4px 8px; text-align: center; border: 1px solid #d9d9d9; font-weight: 500; color: rgba(0,0,0,.85); line-height: 1.2; font-size: 14px; white-space: nowrap; }\
      .my-permission-detail-table td { padding: 4px 8px; border: 1px solid #d9d9d9; vertical-align: middle; line-height: 1.2; }\
      .my-permission-detail-table tbody tr:hover { background-color: rgba(0,0,0,.02); }\
      .my-permission-detail-table-cell { display: flex; align-items: center; flex-wrap: wrap; gap: 4px; }\
      .my-permission-detail-table-cell-content { display: inline-flex; align-items: center; gap: 4px; white-space: nowrap; flex-shrink: 0; }\
      .my-permission-detail-table-cell-empty { color: rgba(0,0,0,.25); font-style: italic; }\
      .my-permission-detail-table-permission-item { display: inline-flex; align-items: center; gap: 4px; margin-right: 8px; white-space: nowrap; flex-shrink: 0; }\
      .my-permission-detail-table-permission-item:last-child { margin-right: 0; }\
      .my-permission-detail-table-name { margin-right: 4px; color: rgba(0,0,0,.85); white-space: nowrap; font-size: 13px; }\
      .my-permission-detail-table-code { padding: 1px 4px; background-color: #f5f5f5; border-radius: 2px; font-size: 11px; color: rgba(0,0,0,.65); font-family: monospace; white-space: nowrap; }\
      .my-permission-detail-table-type { padding: 1px 5px; border-radius: 2px; font-size: 11px; color: #fff; font-weight: 500; white-space: nowrap; }\
      .my-permission-detail-table-type-menu { background-color: #1890ff; }\
      .my-permission-detail-table-type-button { background-color: #fa8c16; }\
      .my-permission-detail-table-type-api { background-color: #52c41a; }\
    ',
  })

  // 注册权限详情表格组件
  NUI.component('my-permission-detail-table', {
    props: {
      permissions: {
        type: Array,
        default: function() {
          return [];
        }
      }
    },
    computed: {
      maxLevel: function() {
        // 计算权限树的最大层级深度
        var max = 0;
        function getMaxLevel(perms, level) {
          if (level > max) {
            max = level;
          }
          if (!perms || perms.length === 0) {
            return;
          }
          for (var i = 0; i < perms.length; i++) {
            if (perms[i].children && perms[i].children.length > 0) {
              getMaxLevel(perms[i].children, level + 1);
            }
          }
        }
        getMaxLevel(this.permissions, 0);
        return max;
      },
      tableHeaders: function() {
        // 生成表头：支持任意层级的权限树
        var headers = [];
        for (var i = 0; i <= this.maxLevel; i++) {
          headers.push('第' + (i + 1) + '级权限');
        }
        return headers;
      },
      tableRows: function() {
        // 将权限树转换为表格行数据，并计算每个单元格的 rowspan
        // 支持任意层级的权限树结构
        var rows = [];
        var maxCols = this.maxLevel + 1;
        
        function countLeafNodes(perm) {
          // 计算某个权限节点下有多少个叶子节点（用于计算 rowspan）
          if (!perm.children || perm.children.length === 0) {
            return 1; // 叶子节点本身
          }
          var count = 0;
          for (var i = 0; i < perm.children.length; i++) {
            count += countLeafNodes(perm.children[i]);
          }
          return count;
        }
        
        function buildRows(perms, level, parentRowData, startRowIndex, outputRows) {
          if (!perms || perms.length === 0) {
            return startRowIndex;
          }
          
          var currentRowIndex = startRowIndex;
          outputRows = outputRows || [];
          
          for (var i = 0; i < perms.length; i++) {
            var perm = perms[i];
            
            // 计算当前权限节点需要合并的行数
            var rowspan = countLeafNodes(perm);
            
            // 记录当前单元格的起始行索引
            var cellStartRow = currentRowIndex;
            
            // 如果有子权限，递归处理
            if (perm.children && perm.children.length > 0) {
              // 构建包含当前层级的父级行数据
              var newParentRowData = new Array(maxCols);
              for (var j = 0; j < level; j++) {
                newParentRowData[j] = parentRowData[j];
              }
              // 设置当前层级（将在第一行设置，其他行标记为跳过）
              newParentRowData[level] = {
                permission: perm,
                rowspan: rowspan,
                startRowIndex: cellStartRow
              };
              
              // 递归处理子权限
              var childRows = [];
              currentRowIndex = buildRows(perm.children, level + 1, newParentRowData, currentRowIndex, childRows);
              
              // 为每个子行添加当前层级的单元格信息
              for (var r = 0; r < childRows.length; r++) {
                var childRow = childRows[r];
                var newRowData = new Array(maxCols);
                
                // 复制子行数据（子行已经包含了level之前的父级信息）
                for (var k = 0; k < maxCols; k++) {
                  newRowData[k] = childRow[k];
                }
                
                // 设置当前层级的单元格（只在第一行设置实际数据，其他行标记为跳过）
                if (r === 0) {
                  newRowData[level] = {
                    permission: perm,
                    rowspan: rowspan,
                    startRowIndex: cellStartRow
                  };
                } else {
                  newRowData[level] = { skip: true }; // 标记为跳过，不渲染
                }
                
                outputRows.push(newRowData);
              }
            } else {
              // 没有子权限，这是一条完整的行
              var newRowData = new Array(maxCols);
              
              // 复制父级行数据
              for (var j = 0; j < level; j++) {
                newRowData[j] = parentRowData[j];
              }
              
              // 设置当前层级的单元格
              newRowData[level] = {
                permission: perm,
                rowspan: rowspan,
                startRowIndex: cellStartRow
              };
              
              // 填充空值
              for (var k = level + 1; k < maxCols; k++) {
                newRowData[k] = null;
              }
              
              outputRows.push(newRowData);
              currentRowIndex++;
            }
          }
          
          return currentRowIndex;
        }
        
        // 从一级权限开始构建
        buildRows(this.permissions, 0, [], 0, rows);
        
        return rows;
      }
    },
    methods: {
      getPermissionTypeText: function(type) {
        var typeMap = {
          1: '菜单',
          2: '按钮',
          3: '接口'
        };
        return typeMap[type] || '未知';
      },
      getPermissionTypeClass: function(type) {
        var classMap = {
          1: 'my-permission-detail-table-type-menu',
          2: 'my-permission-detail-table-type-button',
          3: 'my-permission-detail-table-type-api'
        };
        return classMap[type] || '';
      },
      shouldRenderCell: function(cellData) {
        // 判断某个单元格是否应该渲染（用于 rowspan 合并）
        if (!cellData || cellData === null) {
          return false;
        }
        // 如果标记为跳过，不渲染
        if (typeof cellData === 'object' && cellData.skip === true) {
          return false;
        }
        // 如果有权限数据（单个或多个），需要渲染
        if (typeof cellData === 'object' && (cellData.permission || cellData.permissions)) {
          return true;
        }
        return true;
      },
      getCellRowspan: function(cellData) {
        // 获取单元格的 rowspan 值
        if (!cellData || cellData === null) {
          return 1;
        }
        if (typeof cellData === 'object' && (cellData.permission || cellData.permissions)) {
          return cellData.rowspan || 1;
        }
        return 1;
      },
      getCellPermission: function(cellData) {
        // 从单元格数据中提取权限对象（单个权限）
        if (!cellData || cellData === null) {
          return null;
        }
        if (typeof cellData === 'object' && cellData.permission) {
          return cellData.permission;
        }
        return cellData;
      },
      getCellPermissions: function(cellData) {
        // 从单元格数据中提取权限数组（多个权限，用于合并单元格）
        if (!cellData || cellData === null) {
          return null;
        }
        if (typeof cellData === 'object' && cellData.permissions && Array.isArray(cellData.permissions)) {
          return cellData.permissions;
        }
        // 如果只有单个权限，也返回数组格式
        if (typeof cellData === 'object' && cellData.permission) {
          return [cellData.permission];
        }
        return null;
      },
      isMergedCell: function(cellData) {
        // 判断是否为合并单元格（包含一个或多个权限的数组）
        if (!cellData || cellData === null) {
          return false;
        }
        // 如果有 permissions 数组（无论长度），都认为是合并单元格
        return typeof cellData === 'object' && cellData.permissions && Array.isArray(cellData.permissions) && cellData.permissions.length > 0;
      }
    },
    template: '<div class="my-permission-detail-table-container" style="overflow-x: auto;">' +
      '<table class="my-permission-detail-table">' +
      '<thead>' +
      '<tr>' +
      '<th v-for="(header, index) in tableHeaders" :key="index">{{ header }}</th>' +
      '</tr>' +
      '</thead>' +
      '<tbody>' +
      '<tr v-for="(row, rowIndex) in tableRows" :key="rowIndex">' +
      '<td v-for="(cellData, colIndex) in row" :key="colIndex" v-if="shouldRenderCell(cellData)" :rowspan="getCellRowspan(cellData)">' +
      '<div v-if="isMergedCell(cellData)" class="my-permission-detail-table-cell">' +
      '<div v-for="(perm, permIndex) in getCellPermissions(cellData)" :key="permIndex" class="my-permission-detail-table-permission-item">' +
      '<span class="my-permission-detail-table-name">{{ perm.permission_name }}</span>' +
      '<span v-if="perm.permission_code" class="my-permission-detail-table-code">{{ perm.permission_code }}</span>' +
      '<span v-if="perm.permission_type" :class="[\'my-permission-detail-table-type\', getPermissionTypeClass(perm.permission_type)]">' +
      '{{ getPermissionTypeText(perm.permission_type) }}' +
      '</span>' +
      '</div>' +
      '</div>' +
      '<div v-else-if="getCellPermission(cellData)" class="my-permission-detail-table-cell">' +
      '<div class="my-permission-detail-table-cell-content">' +
      '<span class="my-permission-detail-table-name">{{ getCellPermission(cellData).permission_name }}</span>' +
      '<span v-if="getCellPermission(cellData).permission_code" class="my-permission-detail-table-code">{{ getCellPermission(cellData).permission_code }}</span>' +
      '<span v-if="getCellPermission(cellData).permission_type" :class="[\'my-permission-detail-table-type\', getPermissionTypeClass(getCellPermission(cellData).permission_type)]">' +
      '{{ getPermissionTypeText(getCellPermission(cellData).permission_type) }}' +
      '</span>' +
      '</div>' +
      '</div>' +
      '<div v-else class="my-permission-detail-table-cell my-permission-detail-table-cell-empty">—</div>' +
      '</td>' +
      '</tr>' +
      '</tbody>' +
      '</table>' +
      '</div>'
  });
})()
