<script setup lang='ts'>
import { computed, ref,onMounted,nextTick  } from 'vue'
import { useI18n } from 'vue-i18n'
import {getTemp,AddTemp,UpdateTemp,DelTemp} from "@/api/subcription/temp"
const { t } = useI18n()
interface Temp {
  file: string;
  text: string;
  CreateDate: string;
}
const tableData = ref<Temp[]>([])
const Tempoldname = ref('')
const Tempname = ref('')
const TempText = ref('')
const dialogVisible = ref(false)
const table = ref()
const tempMode = ref<'add' | 'edit'>('add')
const tempTitle = computed(() => t(tempMode.value === 'add' ? 'templates.addTitle' : 'templates.editTitle'))
const radio1 = ref('1')

async function gettemps() {
  const {data} = await getTemp();
    tableData.value = data
}
onMounted(async() => {
  gettemps()
})
const handleAddTemp = ()=>{
  tempMode.value = 'add'
  Tempname.value = ''
  TempText.value = ''
  radio1.value = '1'
  dialogVisible.value = true
}
const addtemp = async ()=>{
   if (tempMode.value === 'add'){
        await AddTemp({
        filename: Tempname.value.trim(),
        text: TempText.value.trim(),
      })
      ElMessage.success(t('templates.addSuccess'));
   }else{
    await UpdateTemp({
        filename: Tempname.value.trim(),
        oldname: Tempoldname.value.trim(),
        text: TempText.value.trim(),
      })
    ElMessage.success(t('templates.updateSuccess'));
   }
    gettemps()
    Tempname.value = ''
    TempText.value = ''
    dialogVisible.value = false;
}

const multipleSelection = ref<Temp[]>([])
const handleSelectionChange = (val: Temp[]) => {
  multipleSelection.value = val
  
}
const selectAll = () => {
    nextTick(() => {
        tableData.value.forEach(row => {
            table.value.toggleRowSelection(row, true)
        })
    })
}
const handleEdit = (row:any) => {
  for (let i = 0; i < tableData.value.length; i++) {
    if (tableData.value[i].file === row.file) {
      tempMode.value = 'edit'
      Tempname.value = tableData.value[i].file
      Tempoldname.value = Tempname.value
      TempText.value = tableData.value[i].text
      dialogVisible.value = true
      // value1.value = tableData.value[i].Nodes.map((item) => item.Name)
    }
  }
}
const toggleSelection = () => {
  table.value.clearSelection()
}

const handleDel = (row:any) => {
  ElMessageBox.confirm(
    t('templates.deletePrompt', { name: row.file }),
    t('common.deletePrompt'),
    {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning',
    }
  ).then(async () => {
      await DelTemp({
        filename: row.file,
        type: row.type
      })
      ElMessage({
        type: 'success',
        message: t('templates.deleteSuccess'),
      })
      gettemps()      
    })
}

const selectDel = () => {
  if (multipleSelection.value.length === 0) {
      return
  }
  ElMessageBox.confirm(
    t('templates.deleteSelectedPrompt'),
    t('common.deletePrompt'),
    {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning',
    }
  ).then( () => {
    for (let i = 0; i < multipleSelection.value.length; i++) {
       DelTemp({
        filename: multipleSelection.value[i].file,
      })
        tableData.value = tableData.value.filter((item) => item.file !== multipleSelection.value[i].file)
      }
      ElMessage({
        type: 'success',
        message: t('templates.deleteSuccess'),
      })
    })

}
// 分页显示
const currentPage = ref(1);
const pageSize = ref(10);
const handleSizeChange = (val: number) => {
  pageSize.value = val;
  // console.log(`每页 ${val} 条`);
}

const handleCurrentChange = (val: number) => {
  currentPage.value = val;
  // console.log(`当前页: ${val}`);
}
// 表格数据静态化
const currentTableData = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value;
  const end = start + pageSize.value;
  return tableData.value.slice(start, end);
});

</script>

<template>
  <div class="management-page">
    <el-dialog
    v-model="dialogVisible"
    :title="tempTitle"
    width="min(820px, 94vw)"
  >
  <el-input 
  v-model="TempText" 
  :placeholder="$t('templates.content')"
  :rows="10" 
  type="textarea" 
  style="margin-bottom: 10px"
  />
  <el-input v-model="Tempname" :placeholder="$t('templates.filename')"/>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="dialogVisible = false">{{ $t('common.close') }}</el-button>
        <el-button type="primary" @click="addtemp">{{ $t('common.confirm') }}</el-button>
      </div>
    </template>
  </el-dialog>
    <el-card class="management-card" shadow="never">
    <div class="management-toolbar">
      <el-button type="primary" @click="handleAddTemp">{{ $t('templates.add') }}</el-button>
    </div>
      <el-table ref="table" :data="currentTableData" style="width: 100%" @selection-change="handleSelectionChange">
    <el-table-column type="selection" fixed prop="ID" label="id"  />
    <el-table-column prop="file" :label="$t('templates.filename')">
      <template #default="scope">
        <el-tag type="success">{{scope.row.file}}</el-tag>
      </template>
    </el-table-column>
    <el-table-column prop="create_date" :label="$t('templates.createdAt')" sortable />
    <el-table-column fixed="right" :label="$t('templates.actions')" width="120">
      <template #default="scope">
        <el-button link type="primary" size="small" @click="handleEdit(scope.row)">{{ $t('common.edit') }}</el-button>
  <el-button link type="primary" size="small" @click="handleDel(scope.row)">{{ $t('common.delete') }}</el-button>

      </template>
    </el-table-column>
  </el-table>
  <div class="management-actions">
    <el-button type="info" @click="selectAll()">{{ $t('common.selectAll') }}</el-button>
    <el-button type="warning" @click="toggleSelection()">{{ $t('common.clearSelection') }}</el-button>
    <el-button type="danger" @click="selectDel">{{ $t('common.deleteSelected') }}</el-button>
  </div>
  <div class="management-footer">
  <el-pagination
  @size-change="handleSizeChange"
  @current-change="handleCurrentChange"
  :current-page="currentPage"
  :page-size="pageSize"
  layout="total, sizes, prev, pager, next, jumper"
  :page-sizes="[10, 20, 30, 40]"
  :total="tableData.length">
</el-pagination>
  </div>

    </el-card>
  </div>
</template>

<style scoped>
.el-input{
  margin-bottom: 10px;
}
</style>@/api/subcription/temp
