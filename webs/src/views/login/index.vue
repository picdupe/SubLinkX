<template>
  <div class="login-container">
    <!-- 顶部 -->
    <div class="absolute-lt flex-x-end p-3 w-full">
      <el-switch
        v-model="isDark"
        inline-prompt
        :active-icon="Moon"
        :inactive-icon="Sunny"
        @change="toggleTheme"
      />
      <lang-select class="ml-2 cursor-pointer" />
    </div>
    <!-- 登录表单 -->
    <el-card class="login-card w-100 <sm:w-85">
      <div class="text-center relative">
        <div class="brand-mark">S</div>
        <p class="login-eyebrow">SECURE SUBSCRIPTION MANAGEMENT</p>
        <h2>{{ defaultSettings.title }}</h2>
        <p class="login-subtitle">{{ $t('login.subtitle') }}</p>
        <el-tag class="version-tag">{{ version }}</el-tag>
      </div>

      <el-form
        ref="loginFormRef"
        :model="loginData"
        :rules="loginRules"
        class="login-form"
      >
        <!-- 用户名 -->
        <el-form-item prop="username">
          <div class="flex-y-center w-full">
            <svg-icon icon-class="user" class="mx-2" />
            <el-input
              ref="usernameInput"
              v-model="loginData.username"
              :placeholder="$t('login.username')"
              name="username"
              size="large"
              class="h-[48px]"
              @keyup.enter.prevent="focusPassword"
            />
          </div>
        </el-form-item>

        <!-- 密码 -->
        <el-tooltip
          :visible="isCapslock"
          :content="$t('login.capsLock')"
          placement="right"
        >
          <el-form-item prop="password">
            <div class="flex-y-center w-full">
              <svg-icon icon-class="lock" class="mx-2" />
              <el-input
                ref="passwordInput"
                v-model="loginData.password"
                :placeholder="$t('login.password')"
                type="password"
                name="password"
                @keyup="checkCapslock"
                @keyup.enter.prevent="focusCaptcha"
                size="large"
                class="h-[48px] pr-2"
                show-password
              />
            </div>
          </el-form-item>
        </el-tooltip>

        <!-- 验证码 -->
        <el-form-item prop="captchaCode">
          <div class="flex-y-center w-full">
            <svg-icon icon-class="captcha" class="mx-2" />
            <el-input
              ref="captchaInput"
              v-model="loginData.captchaCode"
              auto-complete="off"
              size="large"
              class="flex-1"
              :placeholder="$t('login.captchaCode')"
              @keyup.enter.prevent="handleLogin"
            />

            <el-image
              @click="getCaptcha"
              :src="captchaBase64"
              class="cursor-pointer h-[48px]"
            />
          </div>
        </el-form-item>

        <!-- 登录按钮 -->
        <el-button
          :loading="loading"
          type="primary"
          size="large"
          class="w-full"
          @click.prevent="handleLogin"
          >{{ $t("login.login") }}
        </el-button>

      
      </el-form>
    </el-card>

  
  </div>
</template>

<script setup lang="ts">
import { useSettingsStore, useUserStore } from "@/store";
import { getCaptchaApi , GetVersion } from "@/api/auth";
import { LoginData } from "@/api/auth/types";
import { Sunny, Moon } from "@element-plus/icons-vue";
import { LocationQuery, LocationQueryValue, useRoute } from "vue-router";
import router from "@/router";
import defaultSettings from "@/settings";
import { ThemeEnum } from "@/enums/ThemeEnum";
// 获取版本号
const version = ref('')  
const fetchVersion = function(){
  GetVersion().then((res) => {
    console.log("Version fetched:", res.data); // 输出返回内容
    version.value = res.data;
  }).catch((error) => {
    console.error("Error fetching version:", error);
  });
}() 



// Stores
const userStore = useUserStore();
const settingsStore = useSettingsStore();

// Internationalization
const { t } = useI18n();

// Reactive states
const isDark = ref(settingsStore.theme === ThemeEnum.DARK);
const icpVisible = ref(true);
const loading = ref(false); // 按钮loading
const isCapslock = ref(false); // 是否大写锁定
const captchaBase64 = ref(); // 验证码图片Base64字符串
const loginFormRef = ref(ElForm); // 登录表单ref
const passwordInput = ref();
const captchaInput = ref();
const { height } = useWindowSize();

const loginData = ref<LoginData>({
  username: "",
  password: "",
});

function focusPassword() {
  passwordInput.value?.focus();
}

function focusCaptcha() {
  captchaInput.value?.focus();
}

const loginRules = computed(() => {
  return {
    username: [
      {
        required: true,
        trigger: "blur",
        message: t("login.message.username.required"),
      },
    ],
    password: [
      {
        required: true,
        trigger: "blur",
        message: t("login.message.password.required"),
      },
      {
        min: 6,
        message: t("login.message.password.min"),
        trigger: "blur",
      },
    ],
    captchaCode: [
      {
        required: true,
        trigger: "blur",
        message: t("login.message.captchaCode.required"),
      },
    ],
  };
});

/**
 * 获取验证码
 */
function getCaptcha() {
  getCaptchaApi().then(({ data }) => {
    loginData.value.captchaKey = data.captchaKey;
    captchaBase64.value = data.captchaBase64;
  });
}

/**
 * 登录
 */
const route = useRoute();
function handleLogin() {
  loginFormRef.value.validate((valid: boolean) => {
    if (valid) {
      loading.value = true;
      userStore
        .login(loginData.value)
        .then(() => {
          const query: LocationQuery = route.query;
          const redirect = (query.redirect as LocationQueryValue) ?? "/";
          const otherQueryParams = Object.keys(query).reduce(
            (acc: any, cur: string) => {
              if (cur !== "redirect") {
                acc[cur] = query[cur];
              }
              return acc;
            },
            {}
          );

          router.push({ path: redirect, query: otherQueryParams });
        })
        .catch(() => {
          getCaptcha();
        })
        .finally(() => {
          loading.value = false;
        });
    }
  });
}

/**
 * 主题切换
 */

const toggleTheme = () => {
  const newTheme =
    settingsStore.theme === ThemeEnum.DARK ? ThemeEnum.LIGHT : ThemeEnum.DARK;
  settingsStore.changeTheme(newTheme);
};
/**
 * 根据屏幕宽度切换设备模式
 */

watchEffect(() => {
  if (height.value < 600) {
    icpVisible.value = false;
  } else {
    icpVisible.value = true;
  }
});

/**
 * 检查输入大小写
 */
function checkCapslock(event: KeyboardEvent) {
  // 防止浏览器密码自动填充时报错
  if (event instanceof KeyboardEvent) {
    isCapslock.value = event.getModifierState("CapsLock");
  }
}

onMounted(() => {
  getCaptcha();
});
</script>

<style lang="scss" scoped>
.login-container {
	position: relative;
	overflow: hidden;
	background:
		radial-gradient(ellipse at 18% 18%, rgb(147 197 253 / 48%), transparent 34%),
		radial-gradient(ellipse at 82% 78%, rgb(191 219 254 / 55%), transparent 35%),
		linear-gradient(135deg, #edf5ff 0%, #f8fbff 48%, #e7f0ff 100%);
	@apply wh-full flex-center;

	&::before, &::after { position: absolute; width: 360px; height: 360px; border: 1px solid rgb(255 255 255 / 75%); border-radius: 50%; content: ""; pointer-events: none; }
	&::before { top: -170px; right: -100px; box-shadow: 0 0 0 28px rgb(255 255 255 / 13%), 0 0 0 58px rgb(255 255 255 / 10%); }
	&::after { bottom: -230px; left: -120px; width: 450px; height: 450px; box-shadow: 0 0 0 30px rgb(255 255 255 / 15%); }
	> :deep(.absolute-lt) { z-index: 2; }
}

.login-card {
	position: relative;
	z-index: 1;
	width: min(100% - 2rem, 28rem);
	padding: 18px 24px 12px;
	border: 1px solid rgb(255 255 255 / 82%) !important;
	border-radius: var(--app-radius) !important;
	background: rgb(255 255 255 / 66%) !important;
	box-shadow: 0 28px 80px rgb(37 99 235 / 13%), inset 0 1px 0 rgb(255 255 255 / 92%) !important;
	backdrop-filter: blur(24px) saturate(155%);
	-webkit-backdrop-filter: blur(24px) saturate(155%);

	:deep(.el-card__body) { padding: 18px 16px 22px; }
	h2 { margin: 6px 0 0; color: #172033; font-size: 27px; font-weight: 720; letter-spacing: -.04em; }
	.login-eyebrow { margin: 18px 0 0; color: #2488ee; font-size: 10px; font-weight: 750; letter-spacing: .16em; }
	.login-subtitle { margin: 9px 0 0; color: #718096; font-size: 13px; }
	.version-tag { position: absolute; top: 0; right: 0; border-color: #b9dcff; background: #eaf4ff; color: #2488ee; }
	.brand-mark { display: grid; width: 54px; height: 54px; place-items: center; margin: 0 auto; border: 1px solid rgb(255 255 255 / 90%); border-radius: var(--app-radius); background: linear-gradient(145deg, #68b5ff, #409eff); box-shadow: 0 10px 24px rgb(64 158 255 / 28%), inset 0 1px 0 rgb(255 255 255 / 50%); color: white; font-size: 24px; font-weight: 750; }
}

.login-form { padding: 28px 10px 8px; }
.el-form-item { overflow: hidden; margin-bottom: 16px; border: 1px solid rgb(203 213 225 / 76%); border-radius: var(--app-radius); background: rgb(255 255 255 / 62%); transition: border-color 160ms, box-shadow 160ms; }
.el-form-item:focus-within { border-color: #73b8ff; box-shadow: 0 0 0 3px rgb(64 158 255 / 16%); }
.login-form :deep(.el-input__wrapper) { min-height: 48px; padding: 0 12px; background: transparent; box-shadow: none !important; }
.login-form :deep(.el-input__inner) { color: #1e293b; }
.login-form :deep(.el-button--primary) { height: 48px; margin-top: 6px; border: 1px solid rgb(255 255 255 / 48%); border-radius: var(--app-radius); background: linear-gradient(110deg, #58adff, #409eff); box-shadow: 0 8px 18px rgb(64 158 255 / 24%); font-weight: 650; }
.login-form :deep(.el-button--primary:hover) { background: linear-gradient(110deg, #75bcff, #409eff); }
html.dark .login-container { background: radial-gradient(ellipse at 18% 18%, rgb(30 64 110 / 55%), transparent 34%), radial-gradient(ellipse at 82% 78%, rgb(30 58 100 / 45%), transparent 35%), linear-gradient(135deg, #0b1220, #101a2b 55%, #0c1728); }
html.dark .login-card { border-color: rgb(148 163 184 / 22%) !important; background: rgb(15 23 42 / 62%) !important; }
html.dark .login-card h2 { color: #e2e8f0; }
html.dark .el-form-item { border-color: rgb(148 163 184 / 24%); background: rgb(15 23 42 / 34%); }
html.dark .login-form :deep(.el-input__inner) { color: #e2e8f0; }
</style>
