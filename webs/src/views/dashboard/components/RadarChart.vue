<!-- 雷达图 -->
<template>
  <el-card>
    <template #header>{{ $t('charts.orderStatusRadar') }}</template>
    <div :id="id" :class="className" :style="{ height, width }"></div>
  </el-card>
</template>

<script setup lang="ts">
import * as echarts from "echarts";
import { useI18n } from "vue-i18n";
const { t } = useI18n();

const props = defineProps({
  id: {
    type: String,
    default: "radarChart",
  },
  className: {
    type: String,
    default: "",
  },
  width: {
    type: String,
    default: "200px",
    required: true,
  },
  height: {
    type: String,
    default: "200px",
    required: true,
  },
});

const options = {
  grid: {
    left: "2%",
    right: "2%",
    bottom: "10%",
    containLabel: true,
  },
  legend: {
    x: "center",
    y: "bottom",
    data: [t('charts.reserved'), t('charts.ordered'), t('charts.shipped')],
    textStyle: {
      color: "#999",
    },
  },
  radar: {
    // shape: 'circle',
    radius: "60%",
    indicator: [
      { name: t('charts.homeAppliances') },
      { name: t('charts.clothing') },
      { name: t('charts.outdoorSports') },
      { name: t('charts.digitalProducts') },
      { name: t('charts.autoSupplies') },
      { name: t('charts.furniture') },
    ],
  },
  series: [
    {
      name: "Budget vs spending",
      type: "radar",
      itemStyle: {
        borderRadius: 6,
        color: function (params: any) {
          //自定义颜色
          const colorList = ["#409EFF", "#67C23A", "#E6A23C", "#F56C6C"];
          return colorList[params.dataIndex];
        },
      },
      data: [
        {
          value: [400, 400, 400, 400, 400, 400],
          name: t('charts.reserved'),
        },
        {
          value: [300, 300, 300, 300, 300, 300],
          name: t('charts.ordered'),
        },
        {
          value: [200, 200, 200, 200, 200, 200],
          name: t('charts.shipped'),
        },
      ],
    },
  ],
};

const chart = ref<any>("");

onMounted(() => {
  chart.value = markRaw(
    echarts.init(document.getElementById(props.id) as HTMLDivElement)
  );

  chart.value.setOption(options);

  window.addEventListener("resize", () => {
    chart.value.resize();
  });
});

onActivated(() => {
  if (chart.value) {
    chart.value.resize();
  }
});
</script>
