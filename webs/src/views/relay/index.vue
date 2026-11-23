<template>
  <div class="relay-page">
    <el-alert
      :title="t('relay.help')"
      type="info"
      :closable="false"
      show-icon
    />

    <el-tabs v-model="activeTab" class="relay-tabs">
      <el-tab-pane :label="t('relay.serversTab')" name="servers">
        <div class="toolbar">
          <el-button type="primary" @click="openServerDialog()">{{ t('relay.addServer') }}</el-button>
          <el-button @click="loadAll">{{ t('common.refresh') }}</el-button>
        </div>
        <el-table v-loading="loading" :data="servers" stripe table-layout="auto">
          <el-table-column prop="name" :label="t('relay.name')" />
          <el-table-column prop="public_host" :label="t('relay.clientEntry')" />
          <el-table-column prop="agent_id" :label="t('relay.agentId')" />
          <el-table-column prop="haproxy_status" label="HAProxy" />
          <el-table-column :label="t('relay.revision')">
            <template #default="{ row }">{{ row.applied_revision }} / {{ row.desired_revision }}</template>
          </el-table-column>
          <el-table-column :label="t('relay.enabled')">
            <template #default="{ row }"><el-tag :type="row.enabled ? 'success' : 'info'">{{ row.enabled ? t('relay.yes') : t('relay.no') }}</el-tag></template>
          </el-table-column>
          <el-table-column :label="t('relay.actions')">
            <template #default="{ row }">
              <div class="relay-actions">
                <el-button link type="primary" @click="openServerDialog(row as RelayServer)">{{ t('common.edit') }}</el-button>
                <el-button link type="warning" @click="rotateToken(row as RelayServer)">{{ t('relay.rotateToken') }}</el-button>
                <el-button link type="success" @click="reconcile(row as RelayServer)">{{ t('relay.reconcile') }}</el-button>
                <el-button link type="info" @click="showEvents(row as RelayServer)">{{ t('relay.events') }}</el-button>
                <el-button link type="danger" @click="removeServer(row as RelayServer)">{{ t('common.delete') }}</el-button>
              </div>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <el-tab-pane :label="t('relay.routesTab')" name="routes">
        <div class="toolbar">
          <el-button type="primary" :disabled="servers.length === 0 || nodes.length === 0" @click="openRouteDialog()">{{ t('relay.addRoute') }}</el-button>
          <el-button @click="loadAll">{{ t('common.refresh') }}</el-button>
        </div>
        <el-table v-loading="loading" :data="routes" stripe table-layout="auto">
          <el-table-column prop="name" :label="t('relay.routeName')" />
          <el-table-column :label="t('relay.server')">
            <template #default="{ row }">{{ serverName(row.relay_server_id) }}</template>
          </el-table-column>
          <el-table-column :label="t('relay.landingNode')">
            <template #default="{ row }">{{ nodeName(row.landing_node_id) }}</template>
          </el-table-column>
          <el-table-column prop="protocol" :label="t('relay.protocol')" />
          <el-table-column :label="t('relay.listenAddress')">
            <template #default="{ row }">{{ row.listen_address }}:{{ row.listen_port }}</template>
          </el-table-column>
          <el-table-column :label="t('relay.publish')">
            <template #default="{ row }"><el-tag :type="row.publish_enabled ? 'success' : 'info'">{{ row.publish_enabled ? t('relay.allow') : t('relay.off') }}</el-tag></template>
          </el-table-column>
          <el-table-column :label="t('relay.enabled')">
            <template #default="{ row }"><el-switch :model-value="row.enabled" @change="toggleRoute(row as RelayRoute, Boolean($event))" /></template>
          </el-table-column>
          <el-table-column :label="t('relay.actions')">
            <template #default="{ row }">
              <div class="relay-actions">
                <el-button link type="primary" @click="openRouteDialog(row as RelayRoute)">{{ t('common.edit') }}</el-button>
                <el-button link type="danger" @click="removeRoute(row as RelayRoute)">{{ t('common.delete') }}</el-button>
              </div>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>
    </el-tabs>

    <el-dialog v-model="serverDialog" :title="t(editingServer ? 'relay.editServer' : 'relay.addServerTitle')" width="min(520px, 94vw)" destroy-on-close>
      <el-form label-position="top">
        <el-form-item :label="t('relay.name')"><el-input v-model="serverForm.name" maxlength="100" /></el-form-item>
        <el-form-item :label="t('relay.domainOrIp')"><el-input v-model="serverForm.public_host" maxlength="255" /></el-form-item>
        <el-form-item :label="t('relay.enabled')"><el-switch v-model="serverForm.enabled" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="serverDialog = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="saving" @click="saveServer">{{ t('common.save') }}</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="routeDialog" :title="t(editingRoute ? 'relay.editRoute' : 'relay.addRouteTitle')" width="min(560px, 94vw)" destroy-on-close>
      <el-form label-position="top">
        <el-form-item :label="t('relay.routeName')"><el-input v-model="routeForm.name" maxlength="100" /></el-form-item>
        <el-form-item :label="t('relay.server')"><el-select v-model="routeForm.relay_server_id" class="full-width"><el-option v-for="item in servers" :key="item.ID" :label="item.name" :value="item.ID" /></el-select></el-form-item>
        <el-form-item :label="t('relay.landingNode')"><el-select v-model="routeForm.landing_node_id" class="full-width"><el-option v-for="item in nodes" :key="item.ID" :label="item.Name" :value="item.ID" /></el-select></el-form-item>
        <el-form-item :label="t('relay.transport')"><el-select v-model="routeForm.protocol" class="full-width"><el-option :label="t('relay.tcpHaproxy')" value="tcp" /><el-option :label="t('relay.udpAgent')" value="udp" /></el-select></el-form-item>
        <el-row :gutter="14">
          <el-col :span="12"><el-form-item :label="t('relay.listenIp')"><el-input v-model="routeForm.listen_address" /></el-form-item></el-col>
          <el-col :span="12"><el-form-item :label="`${routeForm.protocol.toUpperCase()} ${t('relay.port')}`"><el-input-number v-model="routeForm.listen_port" :min="1" :max="65535" class="full-width" /></el-form-item></el-col>
        </el-row>
        <el-form-item :label="t('relay.publish')"><el-switch v-model="routeForm.publish_enabled" :active-text="t('relay.allow')" :inactive-text="t('relay.off')" /></el-form-item>
        <el-form-item :label="t('relay.priority')"><el-input-number v-model="routeForm.priority" /></el-form-item>
        <el-form-item :label="t('relay.note')"><el-input v-model="routeForm.description" type="textarea" maxlength="1000" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="routeDialog = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="saving" @click="saveRoute">{{ t('common.save') }}</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="eventsDialog" :title="t('relay.latestEvents')" width="min(760px, 94vw)">
      <el-table :data="events" stripe>
        <el-table-column prop="CreatedAt" :label="t('relay.time')" min-width="190" />
        <el-table-column prop="event_type" :label="t('relay.event')" min-width="180" />
        <el-table-column prop="details" :label="t('relay.details')" min-width="240" />
      </el-table>
    </el-dialog>

    <el-dialog v-model="revisionsDialog" :title="t('relay.configHistory')" width="min(820px, 94vw)">
      <el-table :data="revisions" stripe>
        <el-table-column prop="revision" :label="t('relay.version')" width="90" />
        <el-table-column prop="config_sha256" label="SHA-256" min-width="260" />
        <el-table-column prop="change_summary" :label="t('relay.summary')" min-width="180" />
        <el-table-column :label="t('relay.actions')" width="100">
          <template #default="{ row }"><el-button link type="warning" @click="rollback(row)">{{ t('relay.rollback') }}</el-button></template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <el-dialog
      v-model="tokenDialog"
      class="relay-token-dialog"
      :title="t('relay.oneTimeToken')"
      width="min(500px, 92vw)"
      :close-on-click-modal="false"
      @closed="clearAgentToken"
      append-to-body
      destroy-on-close
    >
      <el-alert
        :title="t('relay.tokenWarning')"
        type="warning"
        :closable="false"
        show-icon
      />
      <el-input
        v-model="agentTokenValue"
        class="token-value"
        type="textarea"
        :rows="3"
        readonly
        resize="none"
      />
      <template #footer>
        <el-button @click="tokenDialog = false">{{ t('relay.saved') }}</el-button>
        <el-button type="primary" @click="copyAgentToken">{{ t('relay.copyToken') }}</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="confirmDialog"
      class="relay-confirm-dialog"
      :title="confirmTitle"
      width="min(460px, 92vw)"
      :close-on-click-modal="false"
      @closed="clearConfirmAction"
      append-to-body
    >
      <el-alert
        :title="confirmMessage"
        type="warning"
        :closable="false"
        show-icon
      />
      <template #footer>
        <el-button @click="confirmDialog = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" @click="runConfirmAction">{{ t('common.confirm') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { nextTick, onMounted, reactive, ref } from "vue";
import { useI18n } from "vue-i18n";
import { getNodes } from "@/api/subcription/node";
import {
  createRelayRoute,
  createRelayServer,
  deleteRelayRoute,
  deleteRelayServer,
  listRelayRoutes,
  listRelayEvents,
  listRelayRevisions,
  listRelayServers,
  rotateRelayToken,
  reconcileRelayServer,
  rollbackRelayServer,
  setRelayRouteEnabled,
  updateRelayRoute,
  updateRelayServer,
} from "@/api/relay";

interface RelayServer {
  ID: number;
  name: string;
  public_host: string;
  agent_id: string;
  haproxy_status: string;
  desired_revision: number;
  applied_revision: number;
  enabled: boolean;
}
interface RelayRoute {
  ID: number;
  name: string;
  relay_server_id: number;
  landing_node_id: number;
  protocol: "tcp" | "udp";
  listen_address: string;
  listen_port: number;
  enabled: boolean;
  publish_enabled: boolean;
  priority: number;
  description: string;
}
interface LandingNode { ID: number; Name: string }

const activeTab = ref("servers");
const { t } = useI18n();
const loading = ref(false);
const saving = ref(false);
const servers = ref<RelayServer[]>([]);
const routes = ref<RelayRoute[]>([]);
const nodes = ref<LandingNode[]>([]);
const serverDialog = ref(false);
const routeDialog = ref(false);
const eventsDialog = ref(false);
const revisionsDialog = ref(false);
const tokenDialog = ref(false);
const agentTokenValue = ref("");
const confirmDialog = ref(false);
const confirmTitle = ref("");
const confirmMessage = ref("");
const confirmAction = ref<(() => Promise<void>) | null>(null);
const events = ref<any[]>([]);
const revisions = ref<any[]>([]);
const editingServer = ref<number | null>(null);
const editingRoute = ref<number | null>(null);
const serverForm = reactive({ name: "", public_host: "", enabled: true });
const routeForm = reactive({ name: "", relay_server_id: 0, landing_node_id: 0, protocol: "tcp", listen_address: "0.0.0.0", listen_port: 443, enabled: true, publish_enabled: false, priority: 0, description: "" });

async function loadAll() {
  loading.value = true;
  try {
    const [serverResult, routeResult, nodeResult] = await Promise.all([listRelayServers(), listRelayRoutes(), getNodes()]);
    servers.value = serverResult.data || [];
    routes.value = routeResult.data || [];
    nodes.value = nodeResult.data || [];
  } finally {
    loading.value = false;
  }
}

function serverName(id: number) { return servers.value.find((item) => item.ID === id)?.name || `#${id}`; }
function nodeName(id: number) { return nodes.value.find((item) => item.ID === id)?.Name || `#${id}`; }

function openServerDialog(row?: RelayServer) {
  editingServer.value = row?.ID ?? null;
  Object.assign(serverForm, { name: row?.name || "", public_host: row?.public_host || "", enabled: row?.enabled ?? true });
  serverDialog.value = true;
}

async function saveServer() {
  if (!serverForm.name.trim() || !serverForm.public_host.trim()) return ElMessage.warning(t('relay.fillServer'));
  saving.value = true;
  try {
    const result = editingServer.value
      ? await updateRelayServer(editingServer.value, { ...serverForm })
      : await createRelayServer({ ...serverForm });
    if (!editingServer.value && result.data?.agent_token) {
      agentTokenValue.value = result.data.agent_token;
      tokenDialog.value = true;
    }
    serverDialog.value = false;
    await loadAll();
  } finally { saving.value = false; }
}

function askConfirm(title: string, message: string, action: () => Promise<void>) {
  confirmTitle.value = title;
  confirmMessage.value = message;
  confirmAction.value = action;
  confirmDialog.value = true;
}

async function runConfirmAction() {
  const action = confirmAction.value;
  confirmDialog.value = false;
  confirmAction.value = null;
  await nextTick();
  if (action) await action();
}

function clearConfirmAction() {
  confirmAction.value = null;
}

function clearAgentToken() {
  agentTokenValue.value = "";
}

function rotateToken(row: RelayServer) {
  askConfirm(t('relay.rotateToken'), t('relay.rotatePrompt', { name: row.name }), async () => {
    const result = await rotateRelayToken(row.ID);
    agentTokenValue.value = result.data.agent_token;
    tokenDialog.value = true;
  });
}

async function copyAgentToken() {
  try {
    await navigator.clipboard.writeText(agentTokenValue.value);
    ElMessage.success(t('relay.tokenCopied'));
  } catch {
    ElMessage.error(t('relay.copyTokenFailed'));
  }
}

async function showEvents(row: RelayServer) {
  const result = await listRelayEvents(row.ID);
  events.value = result.data || [];
  eventsDialog.value = true;
}

function reconcile(row: RelayServer) {
  askConfirm(t('relay.syncTitle'), t('relay.syncPrompt', { name: row.name }), async () => {
    await reconcileRelayServer(row.ID);
    ElMessage.success(t('relay.taskCreated'));
    const result = await listRelayRevisions(row.ID);
    revisions.value = result.data || [];
    revisionsDialog.value = true;
  });
}

function rollback(row: any) {
  askConfirm(t('relay.rollbackTitle'), t('relay.rollbackPrompt', { version: row.revision }), async () => {
    await rollbackRelayServer(row.relay_server_id, row.revision);
    ElMessage.success(t('relay.rollbackTaskCreated'));
    revisionsDialog.value = false;
    await loadAll();
  });
}

function removeServer(row: RelayServer) {
  askConfirm(t('relay.deleteServerTitle'), t('relay.deleteServerPrompt', { name: row.name }), async () => {
    await deleteRelayServer(row.ID);
    await loadAll();
  });
}

function openRouteDialog(row?: RelayRoute) {
  editingRoute.value = row?.ID ?? null;
  Object.assign(routeForm, row ? { ...row } : { name: "", relay_server_id: servers.value[0]?.ID || 0, landing_node_id: nodes.value[0]?.ID || 0, protocol: "tcp", listen_address: "0.0.0.0", listen_port: 443, enabled: true, publish_enabled: false, priority: 0, description: "" });
  routeDialog.value = true;
}

async function saveRoute() {
  if (!routeForm.name.trim() || !routeForm.relay_server_id || !routeForm.landing_node_id || !routeForm.listen_port) return ElMessage.warning(t('relay.fillRoute'));
  saving.value = true;
  try {
    if (editingRoute.value) await updateRelayRoute(editingRoute.value, { ...routeForm });
    else await createRelayRoute({ ...routeForm });
    routeDialog.value = false;
    await loadAll();
  } finally { saving.value = false; }
}

async function toggleRoute(row: RelayRoute, enabled: boolean) {
  await setRelayRouteEnabled(row.ID, enabled);
  await loadAll();
}

function removeRoute(row: RelayRoute) {
  askConfirm(t('relay.deleteRouteTitle'), t('relay.deleteRoutePrompt', { name: row.name }), async () => {
    await deleteRelayRoute(row.ID);
    await loadAll();
  });
}

onMounted(loadAll);
</script>

<style scoped>
.relay-page { display: flex; flex-direction: column; gap: 16px; min-width: 0; }
.relay-tabs { padding: 16px; border-radius: 8px; background: var(--el-bg-color); }
.toolbar { display: flex; justify-content: flex-end; gap: 8px; margin-bottom: 14px; }
.relay-actions { display: flex; flex-wrap: wrap; gap: 0 8px; }
.relay-actions :deep(.el-button) { margin-left: 0; }
.relay-token-dialog { --el-dialog-bg-color: var(--el-bg-color-overlay); }
.relay-confirm-dialog { --el-dialog-bg-color: var(--el-bg-color-overlay); }
.token-value { margin-top: 16px; }
.token-value :deep(textarea) { font-family: monospace; overflow-wrap: anywhere; }
.full-width { width: 100%; }
</style>
