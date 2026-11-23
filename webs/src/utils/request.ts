import axios, { InternalAxiosRequestConfig, AxiosResponse } from "axios";
import { useUserStoreHook } from "@/store/modules/user";
import i18n from "@/lang";

const t = (key: string) => i18n.global.t(key);
const localizedResponseMessage = (message?: string) => {
  if (!message || /[\u4e00-\u9fff]/.test(message)) return t("errors.request");
  return message;
};

// 创建 axios 实例
const service = axios.create({
  baseURL: import.meta.env.VITE_APP_BASE_API,
  timeout: 50000,
  headers: { "Content-Type": "application/json;charset=utf-8" },
});

// 请求拦截器
service.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const accessToken = localStorage.getItem("accessToken");
    if (accessToken) {
      config.headers.Authorization = accessToken;
    }
    return config;
  },
  (error: any) => {
    return Promise.reject(error);
  }
);

// 响应拦截器
service.interceptors.response.use(
  (response: AxiosResponse) => {
    const { code, msg } = response.data;
    if (code === "00000") {
      return response.data;
    }
    // 响应数据为二进制流处理(Excel导出)
    if (response.data instanceof ArrayBuffer) {
      return response;
    }

    const message = localizedResponseMessage(msg) || t("errors.system");
    ElMessage.error(message);
    return Promise.reject(new Error(message));
  },
  (error: any) => {
    const response = error?.response;
    if (!response) {
      const message = t("errors.network");
      ElMessage.error(message);
      return Promise.reject(error?.message || error);
    }
    const { code, msg } = response.data || {};
    // 后端对过期或签名错误的 JWT 返回 HTTP 401。
    if (response.status === 401 || code === "A0230" || code === 401) {
      localStorage.removeItem("accessToken");
      const userStore = useUserStoreHook();
      userStore.resetToken().finally(() => {
        if (location.pathname !== "/login") {
          location.href = "/login";
        }
      });
    } else {
      ElMessage.error(localizedResponseMessage(msg) || t("errors.system"));
    }
    return Promise.reject(error.message);
  }
);

// 导出 axios 实例
export default service;
