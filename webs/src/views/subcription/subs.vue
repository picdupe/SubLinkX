<script setup lang='ts'>
import { computed, ref, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import {getSubs,AddSub,DelSub,UpdateSub,rotateSubToken} from "@/api/subcription/subs"
import {getTemp} from "@/api/subcription/temp"
import {getNodes} from "@/api/subcription/node"
import { listRelayRoutes, listRelayServers } from "@/api/relay"
import QrcodeVue from 'qrcode.vue'
import { VueDraggable } from 'vue-draggable-plus'

interface Sub {
  ID: number;
  Name: string;
  CreateDate: string;
  Config: Config;
  Nodes: Node[];
  RelayRoutes?: RelayRoute[];
  SubLogs:SubLogs[];
}
interface RelayRoute {
  ID: number;
  name: string;
  relay_server_id: number;
  landing_node_id: number;
  enabled: boolean;
  publish_enabled: boolean;
  listen_address: string;
  listen_port: number;
}
interface RelayServer {
  ID: number;
  name: string;
  public_host: string;
  enabled: boolean;
}
interface Node {
  ID: number;
  Name: string;
  Link: string;
  CreateDate: string;
}
interface Config {
  clash: string;
  surge:string;
  udp: boolean;
  cert: boolean;
  support_x25519mlkem768: boolean;
}
interface SubLogs {
  date: string;
  name: string;
  count: number;
  address: string;
}
interface Temp {
  file: string;
  text: string;
  CreateDate: string;
}
const { t } = useI18n()
const tableData = ref<Sub[]>([])
const Clash = ref('')
const Surge = ref('')
const subMode = ref<'add' | 'edit'>('add')
const subTitle = computed(() => t(subMode.value === 'add' ? 'subscriptions.addTitle' : 'subscriptions.editTitle'))
const Subname = ref('')
const oldSubname = ref('')
const dialogVisible = ref(false)
const table = ref()
const NodesList = ref<Node[]>([])
const relayRouteList = ref<RelayRoute[]>([])
const relayServerList = ref<RelayServer[]>([])
const value1 = ref<string[]>([])
const selectedRelayRouteIDs = ref<number[]>([])
const selectedNodeIDs = computed(() => new Set(
  NodesList.value.filter((item) => value1.value.includes(item.Name)).map((item) => item.ID)
))
const canConfigureRealityKem = computed(() => NodesList.value.some((item) => {
  if (!value1.value.includes(item.Name)) return false
  try {
    const uri = new URL(item.Link.trim())
    return uri.protocol === 'vless:' && uri.searchParams.get('security')?.toLowerCase() === 'reality'
  } catch {
    return /^vless:\/\//i.test(item.Link) && /(?:\?|&)security=reality(?:&|#|$)/i.test(item.Link)
  }
}))
const availableRelayRoutes = computed(() => {
  const enabledServers = new Set(relayServerList.value.filter((item) => item.enabled).map((item) => item.ID))
  return relayRouteList.value.filter((item) =>
    item.enabled && item.publish_enabled && enabledServers.has(item.relay_server_id) && selectedNodeIDs.value.has(item.landing_node_id)
  )
})
watch(value1, () => {
  const eligibleIDs = new Set(availableRelayRoutes.value.map((item) => item.ID))
  selectedRelayRouteIDs.value = selectedRelayRouteIDs.value.filter((id) => eligibleIDs.has(id))
}, { deep: true })
watch(canConfigureRealityKem, (canConfigure) => {
  if (!canConfigure) {
    checkList.value = checkList.value.filter((item) => item !== 'support_x25519mlkem768')
  }
}, { flush: 'post' })
function relayRouteLabel(route: RelayRoute) {
  const server = relayServerList.value.find((item) => item.ID === route.relay_server_id)
  return `${route.name} — ${server?.name || `#${route.relay_server_id}`} (${server?.public_host || t('subscriptions.unknownEndpoint')}:${route.listen_port})`
}
const checkList = ref<string[]>([]) // 配置列表
const iplogsdialog = ref(false)
const IplogsList = ref<SubLogs[]>([])
const qrcode = ref('')
const templist = ref<Temp[]>([])
async function getsubs() {
  const {data} = await getSubs();
    tableData.value = data
}
async function gettemps() {
    const {data} = await getTemp();
    templist.value = data
    console.log(templist.value);
}
onMounted(() => {
    getsubs()
    gettemps()
})
onMounted(async() => {
    const {data} = await getNodes();
    NodesList.value = data
})
onMounted(async () => {
  const [routesResult, serversResult] = await Promise.all([listRelayRoutes(), listRelayServers()])
  relayRouteList.value = routesResult.data || []
  relayServerList.value = serversResult.data || []
})


const addSubs = async ()=>{
    const config = JSON.stringify({
    "clash": Clash.value.trim(),
    "surge": Surge.value.trim(),
    "udp": checkList.value.includes('udp') ? true :  false,
    "cert": checkList.value.includes('cert') ? true :  false,
    "support_x25519mlkem768": canConfigureRealityKem.value && checkList.value.includes('support_x25519mlkem768')

  })
  if (subMode.value === 'add') {
    await AddSub({
      config: config,
      name: Subname.value.trim(),
      nodes: value1.value.join(','),
      relay_routes: selectedRelayRouteIDs.value.join(',')
    })
    getsubs()
    ElMessage.success(t('subscriptions.addSuccess'));
  }else{
    await UpdateSub({
      config: config,
      name: Subname.value.trim(),
      nodes: value1.value.join(','),
      relay_routes: selectedRelayRouteIDs.value.join(','),
      oldname: oldSubname.value
    })
    getsubs()
    ElMessage.success(t('subscriptions.updateSuccess'));
  }

    dialogVisible.value = false;
}

const multipleSelection = ref<Sub[]>([])
const handleSelectionChange = (val: Sub[]) => {
  multipleSelection.value = val
  
}
const selectAll = () => {
  tableData.value.forEach(row => {
            table.value.toggleRowSelection(row, true)
        })
}
// IP记录
const handleIplogs = (row: any) => {
  iplogsdialog.value = true
  nextTick(() => {
    tableData.value.forEach((item) => {
    if (item.ID === row.ID) {
      IplogsList.value = item.SubLogs
    }
  })
  
  })
}

const toggleSelection = () => {
  table.value.clearSelection()
}

const handleAddSub = ()=>{
  subMode.value = 'add'
  Subname.value = ''
  oldSubname.value = ''
  checkList.value = []
  Clash.value = './template/clash.yaml'
  Surge.value = './template/surge.conf'
  dialogVisible.value = true
  value1.value = []
  selectedRelayRouteIDs.value = []
}

const handleEdit = (row:any) => {
  for (let i = 0; i < tableData.value.length; i++) {
    if (tableData.value[i].ID === row.ID) {
      function toConfig(value: string | Config): Config {
        if (typeof value === 'string') {
          return JSON.parse(value) as Config;
        } else {
          return value as Config;
        }
      }
      const config = toConfig(tableData.value[i].Config);
      checkList.value = []
      subMode.value = 'edit'
      Subname.value = tableData.value[i].Name
      oldSubname.value = Subname.value
      if (config.udp)  {
        checkList.value.push('udp')
      }
      if (config.cert)  {
        checkList.value.push('cert')
      }
      if (config.support_x25519mlkem768) {
        checkList.value.push('support_x25519mlkem768')
      }
      Clash.value = config.clash
      Surge.value = config.surge
      dialogVisible.value = true
      const selectedNames = new Set(tableData.value[i].Nodes.map((item) => item.Name))
      for (const route of tableData.value[i].RelayRoutes || []) {
        const landingNode = NodesList.value.find((item) => item.ID === route.landing_node_id)
        if (landingNode) selectedNames.add(landingNode.Name)
      }
      value1.value = [...selectedNames]
      const eligibleIDs = new Set(availableRelayRoutes.value.map((item) => item.ID))
      selectedRelayRouteIDs.value = (tableData.value[i].RelayRoutes || [])
        .map((item) => item.ID)
        .filter((id) => eligibleIDs.has(id))
    }
  }
}
const handleDel = (row:any) => {
  ElMessageBox.confirm(
    t('subscriptions.deletePrompt', { name: row.Name }),
    t('common.deletePrompt'),
    {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning',
    }
  ).then(async () => {
      await DelSub({
        id: row.ID
      })
      getsubs()
      ElMessage({
        type: 'success',
        message: t('subscriptions.deleteSuccess'),
      })
      
    })
}

const selectDel = () => {
  if (multipleSelection.value.length === 0) {
      return
  }
  ElMessageBox.confirm(
    t('subscriptions.deleteSelectedPrompt'),
    t('common.deletePrompt'),
    {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning',
    }
  ).then( () => {
    for (let i = 0; i < multipleSelection.value.length; i++) {
       DelSub({
        id: multipleSelection.value[i].ID
      })
        tableData.value = tableData.value.filter((item) => item.ID !== multipleSelection.value[i].ID)
      }
      ElMessage({
        type: 'success',
        message: t('subscriptions.deleteSuccess'),
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
}
// 表格数据静态化
const currentTableData = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value;
  const end = start + pageSize.value;
  return tableData.value.slice(start, end);
});

const formatDateTime = (value: string) => {
  const match = value?.match(/^(\d{4}-\d{2}-\d{2})[T ](\d{2}:\d{2}:\d{2})/);
  return match ? `${match[1]} ${match[2]}` : value;
};

// 复制链接
const copyUrl = (url: string) => {
  const textarea = document.createElement('textarea');
  textarea.value = url;
  document.body.appendChild(textarea);
  textarea.select();
  try {
    const successful = document.execCommand('copy');
    const msg = successful ? 'success' : 'warning';
    const message = successful ? t('subscriptions.copySuccess') : t('subscriptions.copyFailed');
    ElMessage({
      type: msg,
      message,
    });
  } catch (err) {
    ElMessage({
      type: 'warning',
      message: t('subscriptions.copyFailed'),
    });
  } finally {
    document.body.removeChild(textarea);
  }
};

const copyInfo = (row: any) => {
  copyUrl(row.Link)
}
const handleBase64 = (text: string) => {
  return  window.btoa(unescape(encodeURIComponent(text)));
}
const ClientDiaLog = ref(false)
const ClientList = ['v2ray','clash','surge'] // 客户端列表
const ClientUrls = ref<Record<string, string>>({})
const ClientUrl = ref('')
const handleClient = async (row: unknown) => {
  if (!row || typeof row !== 'object' || !('ID' in row) || typeof row.ID !== 'number') return
  try {
    await ElMessageBox.confirm(
      t('subscriptions.rotateTokenPrompt'),
      t('common.confirm'),
      { confirmButtonText: t('common.confirm'), cancelButtonText: t('common.cancel'), type: 'warning' }
    )
    const { data } = await rotateSubToken(row.ID)
    const token = data.token as string
    if (!token) throw new Error('Missing subscription token')
  const serverAddress = import.meta.env.DEV && import.meta.env.VITE_APP_API_URL
    ? import.meta.env.VITE_APP_API_URL.replace(/\/+$/, '')
    : window.location.origin;
  ClientDiaLog.value = true
  ClientUrl.value = `${serverAddress}/c/?token=${encodeURIComponent(token)}`
  ClientList.forEach((item:string) => {
    ClientUrls.value[item]=`${serverAddress}/c/?token=${encodeURIComponent(token)}`
  })
  } catch (error) {
    if (error !== 'cancel' && error !== 'close') throw error
  }
}

const Qrdialog = ref(false)
const QrTitle = ref('')
const handleQrcode = (url:string,title:string)=>{
  Qrdialog.value = true
  qrcode.value = url 
  QrTitle.value = title
}
const OpenUrl = (url:string) => {
  window.open(url)
}
const clientradio = ref('1')

// 注册拖拽函数
const toggleSelect = (name: string) => {
  const index = value1.value.indexOf(name)
  if (index === -1) {
    value1.value.push(name)
  } else {
    value1.value.splice(index, 1)
  }
}

</script>

<template>
  <div class="management-page">
    <el-dialog
      v-model="Qrdialog"
      width="300px"
      style="text-align: center"
      :title="QrTitle"
      :show-close="false"
      :close-on-click-modal="true"
    >
      <div
        class="subscription-qrcode"
      >
        <qrcode-vue
          class="subscription-qrcode-image"
          :value="qrcode"
          :size="200"
          level="H"
          role="button"
          tabindex="0"
          :aria-label="t('subscriptions.qrCopyHint')"
          :title="t('subscriptions.qrCopyHint')"
          @click.stop="copyUrl(qrcode)"
          @keydown.enter.prevent="copyUrl(qrcode)"
          @keydown.space.prevent="copyUrl(qrcode)"
        />
      </div>
      <el-input
        :model-value="qrcode"
        readonly
      >
      </el-input>
      <el-button @click="copyUrl(qrcode)">{{ t('common.copy') }}</el-button>
      <el-button @click="OpenUrl(qrcode)">{{ t('common.open') }}</el-button>
    </el-dialog>

    <el-dialog v-model="ClientDiaLog" :title="t('subscriptions.clientTitle')" style="text-align: center" width="min(560px, 94vw)">
      <el-row>
        <el-col>
        <el-tag type="success" size="large">{{ t('subscriptions.autoDetect') }}</el-tag>
        <el-button @click="handleQrcode(ClientUrl,t('subscriptions.autoDetect'))">{{ t('subscriptions.qrCode') }}</el-button>
      </el-col>
        <el-col v-for="(item,index) in ClientUrls" style="margin-bottom:10px;">
          <el-tag type="success" size="large">{{index}}</el-tag>
          <el-button @click="handleQrcode(`${item}&client=${index}`,index)">{{ t('subscriptions.qrCode') }}</el-button>
        </el-col>
        </el-row>
    </el-dialog>
    
    <el-dialog v-model="iplogsdialog" :title="t('subscriptions.accessLog')" width="min(1000px, 94vw)" draggable>
  <template #footer>
    <div class="dialog-footer">
      <el-table :data="IplogsList" border style="width: 100%">
        <el-table-column prop="IP" label="Ip" />
        <el-table-column prop="Count" :label="t('subscriptions.totalVisits')" />
        <el-table-column prop="Addr" :label="t('subscriptions.source')" />
        <el-table-column prop="Date" :label="t('subscriptions.recentTime')" />
      </el-table>
    </div>
  </template>
</el-dialog>
    <el-dialog
    v-model="dialogVisible"
    :title="subTitle"
    width="min(820px, 94vw)"
  >
  <el-input v-model="Subname" :placeholder="t('subscriptions.namePlaceholder')" />
  
  <el-row >
  <el-tag type="primary">{{ t('subscriptions.templateSelect', { client: 'Clash' }) }}</el-tag>
  <el-radio-group v-model="clientradio" class="ml-4">
      <el-radio value="1">{{ t('subscriptions.local') }}</el-radio>
      <el-radio value="2">{{ t('subscriptions.url') }}</el-radio>
    </el-radio-group>
  <el-select v-model="Clash" :placeholder="t('subscriptions.templateFile')"  v-if="clientradio === '1'">
    <el-option v-for="template in templist" :key="template.file" :label="template.file" :value="'./template/'+template.file" />
  </el-select>
  <el-input v-model="Clash" :placeholder="t('subscriptions.templateFile')"  v-else />
</el-row>
<el-row >
  <el-tag type="primary">{{ t('subscriptions.templateSelect', { client: 'Surge' }) }}</el-tag>
  <el-radio-group v-model="clientradio" class="ml-4">
      <el-radio value="1">{{ t('subscriptions.local') }}</el-radio>
      <el-radio value="2">{{ t('subscriptions.url') }}</el-radio>
    </el-radio-group>
  <el-select v-model="Surge" :placeholder="t('subscriptions.templateFile')"  v-if="clientradio === '1'">
    <el-option v-for="template in templist" :key="template.file" :label="template.file" :value="'./template/'+template.file" />
  </el-select>
  <el-input v-model="Surge" :placeholder="t('subscriptions.templateFile')"  v-else />
</el-row>

  <el-row class="force-options-row">
    <el-tag type="primary">{{ t('subscriptions.forceOptions') }}</el-tag>
  <el-checkbox-group v-model="checkList" class="force-options-group">
    <el-checkbox :value="'udp'">udp</el-checkbox>
    <el-checkbox :value="'cert'">{{ t('subscriptions.skipCertificate') }}</el-checkbox>
    <el-checkbox v-if="canConfigureRealityKem" :value="'support_x25519mlkem768'">{{ t('subscriptions.realityKem') }}</el-checkbox>
  </el-checkbox-group>
</el-row>
  <div class="m-4">
    <p>{{ t('subscriptions.chooseNodes') }}</p>
    <el-select
      v-model="value1"
      multiple
      :placeholder="t('subscriptions.chooseNodePlaceholder')"
      style="width: 100%"
    >
      <el-option
        v-for="item in NodesList"
        :key="item.Name"
        :label="item.Name"
        :value="item.Name"
      />
    </el-select>
    <p>{{ t('subscriptions.selectedNodes') }}</p>
    <VueDraggable v-model="value1" :animation="150" ghost-class="ghost">
      <div v-for="(nodeName, index) in value1" :key="nodeName" class="draggable-item">
        <span class="row-number">{{ index + 1 }}.</span> {{ nodeName }}
      </div>
    </VueDraggable>

    <p>{{ t('subscriptions.chooseRelayRoutes') }}</p>
    <el-select
      v-model="selectedRelayRouteIDs"
      multiple
      clearable
      :placeholder="t('subscriptions.relayPlaceholder')"
      style="width: 100%"
    >
      <el-option
        v-for="route in availableRelayRoutes"
        :key="route.ID"
        :label="relayRouteLabel(route)"
        :value="route.ID"
      />
    </el-select>


  </div>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="dialogVisible = false">{{ t('common.close') }}</el-button>
        <el-button type="primary" @click="addSubs">{{ t('common.confirm') }}</el-button>
      </div>
    </template>
  </el-dialog>
    <el-card class="management-card" shadow="never">
    <div class="management-toolbar">
      <el-button type="primary" @click="handleAddSub">{{ t('subscriptions.addTitle') }}</el-button>
    </div>

      <el-table ref="table"
      :data="currentTableData" 
      :fit="true"
      style="width: 100%" 
      stripe
      @selection-change="handleSelectionChange" 
      row-key="ID" 
      :tree-props="{children: 'Nodes'}"
      >
    <el-table-column type="selection" fixed prop="ID" label="id"  />
    <el-table-column prop="Name" :label="t('subscriptions.nameNode')"  >
    <template #default="{row}">
      <el-tag :type="!row.Nodes ? 'success' : 'primary'" >{{row.Name}}</el-tag>
        </template>
    </el-table-column>
    <el-table-column prop="Link" :label="t('subscriptions.link')" :show-overflow-tooltip="true" >
      <template #default="{row}">
        <div v-if="row.Nodes">
          <el-link type="primary" size="small" @click="handleClient(row)">{{ t('subscriptions.client') }}</el-link>
        </div>
        </template>
      </el-table-column>
 
    <el-table-column prop="CreatedAt" :label="t('subscriptions.createdAt')" sortable>
      <template #default="{ row }">{{ formatDateTime(row.CreatedAt) }}</template>
    </el-table-column>
    <el-table-column :label="t('subscriptions.actions')" fixed="right" width="120">
      <template #default="scope">
        <div v-if="scope.row.Nodes">
          <div class="subscription-row-actions">
            <el-button link type="primary" size="small" @click="handleIplogs(scope.row)">{{ t('subscriptions.logs') }}</el-button>
            <el-button link type="primary" size="small" @click="handleEdit(scope.row)">{{ t('common.edit') }}</el-button>
            <el-button link type="primary" size="small" @click="handleDel(scope.row)">{{ t('common.delete') }}</el-button>
          </div>
        </div>
        <div v-else>
          <el-button link type="primary" size="small" @click="copyInfo(scope.row)">{{ t('common.copy') }}</el-button>
        </div>

      </template>
    </el-table-column>
  </el-table>
  <div class="management-actions">
    <el-button type="info" @click="selectAll()">{{ t('common.selectAll') }}</el-button>
    <el-button type="warning" @click="toggleSelection()">{{ t('common.clearSelection') }}</el-button>
    <el-button type="danger" @click="selectDel">{{ t('common.deleteSelected') }}</el-button>
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
.subscription-qrcode {
  display: flex;
  justify-content: center;
  margin-bottom: 12px;
}

.subscription-qrcode :deep(.subscription-qrcode-image) {
  cursor: pointer;
}

.force-options-row {
  align-items: center;
  gap: 8px;
}

.force-options-group {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  margin: 0;
}

.el-input{
  margin-bottom: 10px;
}
.el-tag{
  margin: 5px;
}
/**拖拽样式 */
.draggable-item {
  padding: 8px 10px;
  margin-bottom: 5px;
  background-color: #f0f2f5;
  border: 1px solid #dcdfe6;
  border-radius: var(--app-radius);
  display: flex;
  align-items: center;
  cursor: grab;
}

.draggable-item:hover {
  background-color: #e6e8eb;
}

.ghost {
  opacity: 0.5;
  background: #c8ebfb;
}

.row-number {
  margin-right: 10px;
  font-weight: bold;
}

.subscription-row-actions {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-wrap: nowrap;
  white-space: nowrap;
}

.subscription-row-actions :deep(.el-button + .el-button) {
  margin-left: 6px;
}
</style>
