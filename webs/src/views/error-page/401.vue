<script setup lang="ts">
import { reactive, toRefs } from "vue";
import { useRouter } from "vue-router";

import { defineComponent } from "vue";

defineComponent({
  name: "Page401",
});

const state = reactive({
  errGif: new URL(`../../assets/images/401.gif`, import.meta.url).href,

  ewizardClap:
    "https://wpimg.wallstcn.com/007ef517-bafd-4066-aae4-6883632d9646",
  dialogVisible: false,
});

const { errGif, ewizardClap, dialogVisible } = toRefs(state);

const router = useRouter();

function back() {
  router.back();
}
</script>

<template>
  <div class="errPage-container">
    <el-button icon="el-icon-arrow-left" class="pan-back-btn" @click="back">
      {{ $t('errorPage.back') }}
    </el-button>
    <el-row>
      <el-col :span="12">
        <h1 class="text-jumbo text-ginormous">Oops!</h1>
        {{ $t('errorPage.gifSource') }} <a href="https://zh.airbnb.com/" target="_blank">airbnb</a>
        <h2>{{ $t('errorPage.noPermission') }}</h2>
        <h6>{{ $t('errorPage.contactAdmin') }}</h6>
        <ul class="list-unstyled">
          <li>{{ $t('errorPage.orTry') }}</li>
          <li class="link-type">
            <router-link to="/dashboard">{{ $t('errorPage.home') }}</router-link>
          </li>
          <li class="link-type">
            <a href="https://www.taobao.com/">{{ $t('errorPage.browse') }}</a>
          </li>
          <li>
            <a href="#" @click.prevent="dialogVisible = true">{{ $t('errorPage.viewImage') }}</a>
          </li>
        </ul>
      </el-col>
      <el-col :span="12">
        <img
          :src="errGif"
          width="313"
          height="428"
          alt="Girl has dropped her ice cream."
        />
      </el-col>
    </el-row>
    <el-dialog v-model="dialogVisible" :title="$t('errorPage.preview')">
      <img :src="ewizardClap" class="pan-img" />
    </el-dialog>
  </div>
</template>

<style lang="scss" scoped>
.errPage-container {
  width: 800px;
  max-width: 100%;
  margin: 100px auto;

  .pan-back-btn {
    color: #fff;
    background: #008489;
    border: none !important;
  }

  .pan-gif {
    display: block;
    margin: 0 auto;
  }

  .pan-img {
    display: block;
    width: 100%;
    margin: 0 auto;
  }

  .text-jumbo {
    font-size: 60px;
    font-weight: 700;
    color: #484848;
  }

  .list-unstyled {
    font-size: 14px;

    li {
      padding-bottom: 5px;
    }

    a {
      color: #008489;
      text-decoration: none;

      &:hover {
        text-decoration: underline;
      }
    }
  }
}
</style>
