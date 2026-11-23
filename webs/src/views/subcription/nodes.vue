<script setup lang='ts'>
import { ref,onMounted,nextTick  } from 'vue'
import { useI18n } from 'vue-i18n'
import type { TableInstance } from 'element-plus'
import {getNodes,AddNodes,DelNode,UpdateNode,GetGroup,SetGroup} from "@/api/subcription/node"

const { t } = useI18n()

interface GroupNode {
  ID: number;
  Name: string;
  Nodes :Node[];
}
interface Node {
  ID: number;
  Name: string;
  Link: string;
  CreateDate: string;
  GroupNodes?: GroupNode[]; // 分组信息
  
}
interface NodeInfo {
    ID?: number // 编辑时需要传入ID
    Title?:string 
    Name?: string
    Link: string
    GroupName?: string[] // 分组名称
}
onMounted(async() => {  // 页面开始执行函数
   getnodes()
   GetGroups()
})
const dialogMode = ref<'add' | 'edit'>('add');

// --- 表格选择与操作相关数据 ---
const multipleSelection = ref<Node[]>([]); // Stores selected table items
const multipleTable = ref<TableInstance | null>(null)


const tableRefs = ref<{ [key: string]: any }>({}); // Stores references to each el-table
// --- 表格选择与操作相关数据结束 ---
// const NodeNewLinkInput = ref("")
// const NodeNewNameInput = ref("")
const NodeGroupInput = ref("")
const tableData = ref<Node[]>([])
// 分组列表临时存放数据
const tableDataTemp = ref<Node[]>([])
// 分组列表临时存放数据
const activeName = ref('all')
const Nodedialog = ref (false); // 弹窗是否可见
const Groupdialog = ref (false); // 弹窗是否可见
const NodeForm = ref<NodeInfo>({
    Title: '',
    Name: '',
    Link: '',
    GroupName: [],
  })
const allGroupNames = ref<string[]>([]); // 所有分组名称
const allNodes = ref<string[]>([]); // 所有节点
const nodelistShow = ref(false); // 节点列表
const SelectionNodeGroups = ref<string[]>([]); // 选中的分组
const SelectionNode = ref(''); // 选中的节点

// const SelectionNodes = ref([]); // 选中的节点
const RadioGroup = ref("1"); // 分组单选框
// 将所有输入的值清空
function ClearInput() {
  SelectionNode.value = ''; // 清空选中的节点
  NodeForm.value = { // 清空节点链接输入框
    Title: '',
    Name: '',
    Link: '',
    GroupName: [],
  }
  NodeGroupInput.value = ''; // 清空创建分组输入框
  SelectionNodeGroups.value = []; // 清空选中的分组
  nodelistShow.value = false; // 隐藏节点列表
  Nodedialog.value = false; // 关闭节点添加弹窗
  Groupdialog.value = false; // 关闭分组绑定弹窗
  
}
async function getnodes() {
  const {data} = await getNodes();
  if (data.length > 0) tableDataTemp.value = tableData.value = data
  allNodes.value = []; // 清空 allNodes 数组
  data.forEach((item:any) => {
      allNodes.value.push(item.Name); // 将所有节点添加到 allNodes 中
  });
  
} 
async function GetGroups() {
  const {data} = await GetGroup();
  if (Array.isArray(data) && data.length > 0) {
  allGroupNames.value=data; // 将所有分组名称添加到 allGroupNames 中

}
  RadioGroup.value = allGroupNames.value.length > 0 ? "1" : "2"; // 自动选择单选框值
  // console.log("单选框",RadioGroup.value);
  
}


const handleAddNode = () => {
  dialogMode.value = 'add';
  Nodedialog.value = true;
  NodeForm.value = {
    Title: t('nodes.addTitle'),
    Name: '',
    Link: '',
    GroupName: [],
  };
  SelectionNodeGroups.value = [];
  NodeGroupInput.value = '';
};

const handleEditNode = (row: Node) => {  
  // NodeNewNameInput.value = row.Name; // 编辑时使用原名称
  // NodeNewLinkInput.value = row.Link; // 编辑时使用原链接
  dialogMode.value = 'edit';
  Nodedialog.value = true;
  NodeForm.value = {
    ID: row.ID,
    Title: t('nodes.editTitle'),
    Name: row.Name,
    Link: row.Link,
    GroupName: (row.GroupNodes || []).map(g => g.Name),
  };
  SelectionNodeGroups.value = NodeForm.value.GroupName || [];
  SelectionNode.value = row.Name;
};
const SubmitNodeForm = async (row:any) => {
  const isAdd = dialogMode.value === 'add';
  let links = NodeForm.value.Link.trim().split(/[\n,]/).map(item => item.trim()).filter(item => item);
  if (isAdd && links.length === 0) {
    ElMessage.warning(t('nodes.nodeLinkRequired'));
    return;
  }

  try {
    if (isAdd) {
      for (const link of links) {
        await AddNodes({
          link,
          group: RadioGroup.value === '1' ? SelectionNodeGroups.value.join(',') : NodeGroupInput.value,
        });
      }
      ElMessage.success(t('nodes.addSuccess'));
    } else {
      await UpdateNode({
        id:NodeForm.value.ID,
        name: NodeForm.value.Name, // 新名称
        link: NodeForm.value.Link, // 新链接
        group: RadioGroup.value === '1' ? SelectionNodeGroups.value.join(',') : NodeGroupInput.value,
      });
      ElMessage.success(t('nodes.updateSuccess'));
    }


  } catch (err) {
    ElMessage.error(t('nodes.saveFailed', { action: t(isAdd ? 'nodes.addAction' : 'nodes.updateAction') }));
  }
  getnodes();
  GetGroups();
  ClearInput();
};

// const AddNode = async() => {
//   // 多节点链接输入处理
//   let NodeLinkInputs = NodeNewLinkInput.value.trim().split(/[\n,]/); // 使用换行符或逗号分隔输入的节点链接
//   NodeLinkInputs = NodeLinkInputs.map((item) => item.trim()).filter((item) => item !== ''); // 去除空白和重复的链接
//   if (NodeNewLinkInput.value.trim() === '') {
//     ElMessage.warning('节点链接不能为空');
//     return;
//   }

//   try {
//     // 多节点同步循环添加节点
//     for(const link of NodeLinkInputs) {
//       if (link) {
//           const newNode = {
//           link: link.trim(), // 节点链接
//           group: SelectionNodeGroups.value.join(','), // 选中的分组
//           };
//           await AddNodes(newNode).then(() => {
//           ElMessage.success('节点添加成功');
//           Nodedialog.value = false; // 关闭弹窗
//           });
//       }
//     }
//     // getnodes(); // 刷新节点列表
//     // GetGroups(); // 刷新分组列表
//   } catch (error) {
//     console.error('添加节点失败:', error);
//     // ElMessage.error('添加节点失败，请稍后再试');
//   }
//   getnodes(); // 刷新节点列表
//   GetGroups(); // 刷新分组列表
//   ClearInput(); // 清空所有输入
// }
const AddGroup = async() => {
  console.log(SelectionNode.value);

  try {
    // 检查是否选择了已有分组或输入了新分组名
    console.log(RadioGroup.value, SelectionNodeGroups.value, NodeGroupInput.value);
    
    if (RadioGroup.value === "1" && SelectionNodeGroups.value.length === 0) {
      ElMessage.warning(t('nodes.noGroupsSelected'));
      return;
    }
    if (RadioGroup.value === "2"&&NodeGroupInput.value.trim() === '') {
      ElMessage.warning(t('nodes.groupNameRequired'));
      return;
  }
      if (SelectionNode.value.length > 0) { // 如果没有选择节点
      const newNode = {
      name: SelectionNode.value, // 节点链接
      group: RadioGroup.value == '1' ?SelectionNodeGroups.value.join(','):NodeGroupInput.value, // 条件选择已有节点|创建分组
      };
      await SetGroup(newNode).then(() => {
      ElMessage.success(t('nodes.groupBindSuccess'));
            });
    }
  } catch (error) {
    console.error('Failed to add group:', error);
    // ElMessage.error('添加分组失败');
  }
  getnodes(); // 刷新节点列表
  GetGroups(); // 刷新分组列表
  ClearInput(); // 清空所有输入
}
// 表格时间格式化
const Timeformatter  = (row:any)=>{
  row.CreatedAt = new Date(row.CreatedAt).toLocaleString(); // 转换为本地时间字符串
  return row.CreatedAt;
  
}
// 选择已有节点显示所属分组
const  handleShownodeGroupList =()=>{
  // 显示这个节点关联的分组
  const nodeData = allNodes.value.find(node => node === SelectionNode.value);
  SelectionNodeGroups.value = []
  tableData.value.forEach((item, ) => {
    if (item.Name === nodeData && (item.GroupNodes?.length ?? 0) > 0) {
      // console.log(`节点 ${nodeData} 的分组:`, item.GroupNodes);
      item.GroupNodes?.forEach((item) => {
        SelectionNodeGroups.value.push(item.Name); // 将分组名称添加到 SelectionNodeGroups 中
      });
    } 
});
}
// 表格所属分组格式化
const Groupformatter = (row:any,cellValue:any) =>{
  const data = row.GroupNodes || [];
  if (!Array.isArray(data) || data.length === 0) {
    return t('nodes.ungrouped'); // 如果没有分组，返回默认值
  }
 return data.map((group: any) => group.Name).join(', ');
}
// --- 复制链接 (保持不变) ---
const copyUrl = (url: string) => {
  if (navigator.clipboard) {
    navigator.clipboard.writeText(url)
      .then(() => {
        ElMessage.success(t('nodes.copied'));
      })
      .catch(err => {
        console.error('Copy failed:', err);
        ElMessage.error(t('nodes.copyFailed'));
      });
  } else {
    const textarea = document.createElement('textarea');
    textarea.value = url;
    document.body.appendChild(textarea);
    textarea.select();
    try {
      document.execCommand('copy');
      ElMessage.success(t('nodes.copied'));
    } catch (err) {
      ElMessage.warning(t('nodes.copyFailed'));
    } finally {
      document.body.removeChild(textarea);
    }
  }
};
// 复制表格节点信息
const copyInfo = (row: Node) => {
  copyUrl(row.Link);
};
const handleDel = async (row: Node) => {
  try {
    await ElMessageBox.confirm(
      t('nodes.deletePrompt', { name: row.Name }),
      t('common.deletePrompt'),
      {
        confirmButtonText: t('common.confirm'),
        cancelButtonText: t('common.cancel'),
        type: 'warning',
      }
    );
    await DelNode({ id: row.ID });
    ElMessage.success(t('nodes.deleteSuccess'));
  } catch (error) {
    if (error !== 'cancel') {
      console.error("Delete failed:", error);
      ElMessage.error(t('nodes.deleteFailed'));
    }
  }
  // 刷新节点列表
  await GetGroups(); // 刷新分组列表
  await getnodes(); // 刷新节点列表
  ClearInput(); // 清空所有输入
};
const selectDel = async () => {
  
  if (multipleSelection.value.length === 0) {
    ElMessage.warning(t('nodes.selectDelete'));
    return;
  }
  try {
    await ElMessageBox.confirm(
      t('nodes.deleteSelectedPrompt', { count: multipleSelection.value.length }),
      t('common.deletePrompt'),
      {
        confirmButtonText: t('common.confirm'),
        cancelButtonText: t('common.cancel'),
        type: 'warning',
      }
    );

    const IDs: number[] = []

    for (const item of multipleSelection.value) {
      await DelNode({ id: item.ID });
       IDs.push(item.ID); // 收集所有已删除的节点ID
    }
    ElMessage.success(t('nodes.batchDeleteSuccess'));
    // 从 tableData 中删除已删除的节点
    tableData.value = tableData.value.filter(item => !IDs.includes(item.ID));
; 

  } catch (error) {
    if (error !== 'cancel') {
      console.error("Batch delete failed:", error);
      ElMessage.error(t('nodes.batchDeleteFailed'));
    }
  }
    // 刷新节点列表
  await GetGroups(); // 刷新分组列表
  await getnodes();
};
// 全选
const selectAll = () => {
  nextTick(() => {
const table = multipleTable.value
  if (table) {
      // 否则全选
      tableData.value.forEach(row => {
        table.toggleRowSelection(row, true)
      })
  }
  });
};
// 取消全选 
const selectClear = () => {
  nextTick(() => {
    const table = multipleTable.value;
    if (table) {
      table.clearSelection();
    }
  });
};
// --- 表格选择操作 (保持不变) ---
const setTableRef = (el: any, name: string) => {
  if (el) {
    tableRefs.value[name] = el;
  } else {
    delete tableRefs.value[name];
  }
};
//批量复制
const selectCopy = async () => {
  if (multipleSelection.value.length === 0) {
    ElMessage.warning(t('nodes.selectCopy'));
    return;
  }
  try {
    copyUrl(multipleSelection.value.map(item => item.Link).join('\n'));
  } catch (error) {
    if (error !== 'cancel') {
      console.error("Batch copy failed:", error);
      ElMessage.error(t('nodes.batchCopyFailed'));
    }
  }
};
const handleSelectionChange = (val: Node[]) => {
  multipleSelection.value = val;
};

watch(activeName, (newVal) => {
  if (newVal === 'all') {
    tableData.value = tableDataTemp.value;
  } else {
    tableData.value = tableDataTemp.value.filter(item => {
      return item.GroupNodes?.some(group => group.Name === newVal);
    });
  }
});


</script>

<template>
  <div class="management-page">
 <el-dialog v-model="Nodedialog" :title="NodeForm.Title" width="min(720px, 94vw)">
  <el-input
    v-model="NodeForm.Link"
    :placeholder="$t('nodes.linkPlaceholder')"
    type="textarea"
    style="margin-bottom: 10px"
    :autosize="{ minRows: 2, maxRows: 10 }"
    v-if="dialogMode === 'add'"
  />

<el-input
  v-model="NodeForm.Name"
  :placeholder="$t('nodes.namePlaceholder')"
  style="margin-bottom: 10px"
  v-if="dialogMode === 'edit'"
/>
  <el-input
    v-model="NodeForm.Link"
    :placeholder="$t('nodes.linkPlaceholder')"
    type="textarea"
    style="margin-bottom: 10px"
    :autosize="{ minRows: 2, maxRows: 10 }"
    v-if="dialogMode === 'edit'"
  />

  <!-- 分组部分 -->
  <el-radio v-model="RadioGroup" label="1" v-if="allGroupNames.length > 0">{{ $t('nodes.chooseGroups') }}</el-radio>
  <el-radio v-model="RadioGroup" label="2">{{ $t('nodes.createGroup') }}</el-radio>

  <div v-if="RadioGroup === '1' && allGroupNames.length > 0">
    <el-select v-model="SelectionNodeGroups" multiple :placeholder="$t('nodes.groupPlaceholder')" class="default">
      <el-option v-for="item in allGroupNames" :key="item" :label="item" :value="item" />
    </el-select>
  </div>

  <el-input v-if="RadioGroup === '2'" v-model="NodeGroupInput" :placeholder="$t('nodes.groupNamePlaceholder')" class="default" />

  <el-button type="primary" @click="SubmitNodeForm">{{ $t(dialogMode === 'add' ? 'nodes.addAction' : 'nodes.updateAction') }}</el-button>
  <el-button @click="Nodedialog = false">{{ $t('common.cancel') }}</el-button>
</el-dialog>


  <!-- 显示表格数据 -->
  <el-card class="management-card" shadow="never">
    <el-tabs v-model="activeName" >
      <el-tab-pane :label="$t('nodes.all') + `(${allNodes.length})`" name="all" />
      <el-tab-pane :label="item" :name="item" v-for="item in allGroupNames" :key="item" />
    </el-tabs>
      <div class="management-toolbar">
        <el-button type="primary" @click="handleAddNode">{{ $t('nodes.add') }}</el-button>
      </div>
      <div style="margin-bottom: 10px"></div>
      <el-table
      ref="multipleTable"
    :data="tableData"
    :fit="true"
    tooltip-effect="dark"
    stripe
    style="width: 100%"
    row-key="ID" 
    :tree-props="{children: 'Nodes'}"
    @selection-change="handleSelectionChange"
    >
        <el-table-column
      type="selection"
      fixed>
    </el-table-column>

    <el-table-column
      type="index"
      >
    </el-table-column>
    <el-table-column
      prop="Name"
      :label="$t('nodes.name')"
      sortable
      >
    <template #default="{row}">
      <el-tag effect="plain" >{{row.Name}}</el-tag>
        </template>
    </el-table-column>
    <el-table-column
      prop="Link"
      :label="$t('nodes.link')"
      :show-overflow-tooltip="true"
      >
          <template #default="{row}">
      <el-tag class="node-link-tag" effect="plain" type="success">{{row.Link}}</el-tag>
        </template>
    </el-table-column>
    
        <el-table-column
      prop="CreatedAt"
      :label="$t('nodes.createdAt')"
      :formatter="Timeformatter"
      sortable
      show-overflow-tooltip>
    </el-table-column>
            <el-table-column
      :label="$t('nodes.groups')"
      :formatter="Groupformatter"
      show-overflow-tooltip>
    </el-table-column>
                <el-table-column :label="$t('nodes.actions')" fixed="right" width="120">
              <template #default="scope">
                <div class="node-row-actions">
                  <el-button link type="primary" size="small" @click="handleEditNode(scope.row as Node)">{{ $t('common.edit') }}</el-button>
                  <el-button link type="primary" size="small" @click="copyInfo(scope.row as Node)">{{ $t('common.copy') }}</el-button>
                  <el-button link type="primary" size="small" @click="handleDel(scope.row as Node)">{{ $t('common.delete') }}</el-button>
                </div>
              </template>
            </el-table-column>
  </el-table>
   <div class="management-actions">
   <el-button type="info" @click="selectAll">{{ $t('common.selectAll') }}</el-button>
   <el-button type="warning" @click="selectClear">{{ $t('common.clearSelection') }}</el-button>
      <el-button type="primary" @click="selectCopy">{{ $t('common.copySelected') }}</el-button>
      <el-button type="danger" @click="selectDel">{{ $t('common.deleteSelected') }}</el-button>
      </div>
  </el-card>
  <!-- 显示表格数据结束 -->
  </div>
</template>
<style>
 /* 创建默认样式 */
 .default {
  font-size: 14px;
  color: #333;
  line-height: 1.6;
  margin-bottom: 10px;
 }
</style>
<style scoped>
.management-page .node-link-tag {
  box-sizing: border-box;
  max-width: 100%;
  overflow: hidden;
  vertical-align: middle;
}

.management-page .node-link-tag :deep(.el-tag__content) {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.management-page .node-row-actions {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-wrap: nowrap;
  white-space: nowrap;
}

.management-page .node-row-actions :deep(.el-button + .el-button) {
  margin-left: 8px;
}
</style>
