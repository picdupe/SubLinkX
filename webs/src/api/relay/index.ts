import request from "@/utils/request";

const base = "/api/v1/admin/relay";

export const listRelayServers = () => request({ url: `${base}/servers`, method: "get" });
export const createRelayServer = (data: unknown) => request({ url: `${base}/servers`, method: "post", data });
export const updateRelayServer = (id: number, data: unknown) => request({ url: `${base}/servers/${id}`, method: "put", data });
export const deleteRelayServer = (id: number) => request({ url: `${base}/servers/${id}`, method: "delete" });
export const rotateRelayToken = (id: number) => request({ url: `${base}/servers/${id}/rotate-token`, method: "post" });
export const listRelayEvents = (id: number) => request({ url: `${base}/servers/${id}/events`, method: "get" });
export const reconcileRelayServer = (id: number) => request({ url: `${base}/servers/${id}/reconcile`, method: "post", data: {} });
export const listRelayRevisions = (id: number) => request({ url: `${base}/servers/${id}/revisions`, method: "get" });
export const rollbackRelayServer = (id: number, revision: number) => request({ url: `${base}/servers/${id}/rollback`, method: "post", data: { revision } });

export const listRelayRoutes = () => request({ url: `${base}/routes`, method: "get" });
export const createRelayRoute = (data: unknown) => request({ url: `${base}/routes`, method: "post", data });
export const updateRelayRoute = (id: number, data: unknown) => request({ url: `${base}/routes/${id}`, method: "put", data });
export const deleteRelayRoute = (id: number) => request({ url: `${base}/routes/${id}`, method: "delete" });
export const setRelayRouteEnabled = (id: number, enabled: boolean) =>
  request({ url: `${base}/routes/${id}/${enabled ? "enable" : "disable"}`, method: "post" });

