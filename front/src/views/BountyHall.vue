<script setup>
import { ref } from 'vue'
import BountyCard from '@/components/BountyCard.vue'

// 搜索关键词
const searchKeyword = ref('')

// 筛选条件 - 下拉框选项
const fabricTypeOptions = [
  { label: '全部', value: '' },
  { label: '棉布', value: 'cotton' },
  { label: '涤纶', value: 'polyester' },
  { label: '尼龙', value: 'nylon' },
  { label: '丝绸', value: 'silk' },
  { label: '麻布', value: 'linen' }
]
const selectedFabricType = ref('')

const priceRangeOptions = [
  { label: '全部', value: '' },
  { label: '1万以下', value: '0-10000' },
  { label: '1-5万', value: '10000-50000' },
  { label: '5-10万', value: '50000-100000' },
  { label: '10万以上', value: '100000+' }
]
const selectedPriceRange = ref('')

const craftOptions = [
  { label: '全部', value: '' },
  { label: '染色', value: 'dye' },
  { label: '印花', value: 'print' },
  { label: '提花', value: 'jacquard' },
  { label: '刺绣', value: 'embroidery' },
  { label: '涂层', value: 'coating' },
  { label: '复合', value: 'composite' }
]
const selectedCraft = ref('')

// 静态数据，后续替换为API
const tasks = ref([
  {
    id: 1,
    title: '急需全棉梭织布，用于夏季服装生产',
    publishTime: '2026-01-16 10:30',
    tags: ['克重：150g/m²', '门幅：150cm', '工艺：梭织', '数量：500m', '成分：全棉'],
    description: '寻找优质全棉梭织布供应商，用于夏季衬衫生产。需要提供布样确认，要求布面光洁、无明显疵点，色牢度4级以上。有OEKO-TEX认证的供应商优先。',
    deadline: '2026-02-01'
  },
  {
    id: 2,
    title: '涤纶面料采购，用于运动服生产',
    publishTime: '2026-01-15 09:15',
    tags: ['克重：120g/m²', '门幅：160cm', '工艺：针织', '数量：1000m', '成分：涤纶'],
    description: '需要速干透气的涤纶面料，用于运动T恤生产。要求面料手感柔软，吸湿排汗性能好。',
    deadline: '2026-02-15'
  },
  {
    id: 3,
    title: '真丝面料寻源，高端女装定制',
    publishTime: '2026-01-14 16:20',
    tags: ['克重：80g/m²', '门幅：140cm', '工艺：印花', '数量：200m', '成分：100%桑蚕丝'],
    description: '高端女装品牌寻找优质真丝面料，要求丝滑柔软，光泽度好，适合制作连衣裙和衬衫。',
    deadline: '2026-01-30'
  }
])

const handleSearch = () => {
  // TODO: 接入远程搜索API
  console.log('搜索:', searchKeyword.value, selectedFabricType.value, selectedPriceRange.value, selectedCraft.value)
}
</script>

<template>
  <div class="w-[80%] mx-auto flex gap-6">
    <!-- 左侧：搜索和筛选区域 -->
    <aside class="w-[30%] shrink-0">
      <div class="bg-white rounded-lg p-8 shadow-sm sticky top-24">
        <!-- 搜索框 -->
        <div class="relative mb-6">
          <input
            v-model="searchKeyword"
            type="text"
            placeholder="搜索悬赏任务..."
            class="w-full px-4 py-2.5 pl-10 border border-gray-200 rounded-lg text-sm focus:outline-none focus:border-blue-500"
            @keyup.enter="handleSearch"
          >
          <i class="fas fa-search absolute left-4 top-1/2 -translate-y-1/2 text-gray-400 text-sm"></i>
        </div>

        <!-- 筛选条件 -->
        <div class="flex flex-col gap-5">
          <!-- 布料类型 -->
          <div class="flex items-center justify-between gap-3">
            <span class="text-sm text-gray-600 whitespace-nowrap">布料类型</span>
            <select
              v-model="selectedFabricType"
              class="px-3 py-1.5 border border-gray-200 rounded text-sm text-gray-600 focus:outline-none focus:border-blue-500 flex-1"
            >
              <option v-for="opt in fabricTypeOptions" :key="opt.value" :value="opt.value">
                {{ opt.label }}
              </option>
            </select>
          </div>

          <!-- 悬赏金额 -->
          <div class="flex items-center justify-between gap-3">
            <span class="text-sm text-gray-600 whitespace-nowrap">悬赏金额</span>
            <select
              v-model="selectedPriceRange"
              class="px-3 py-1.5 border border-gray-200 rounded text-sm text-gray-600 focus:outline-none focus:border-blue-500 flex-1"
            >
              <option v-for="opt in priceRangeOptions" :key="opt.value" :value="opt.value">
                {{ opt.label }}
              </option>
            </select>
          </div>

          <!-- 面料工艺 -->
          <div class="flex items-center justify-between gap-3">
            <span class="text-sm text-gray-600 whitespace-nowrap">面料工艺</span>
            <select
              v-model="selectedCraft"
              class="px-3 py-1.5 border border-gray-200 rounded text-sm text-gray-600 focus:outline-none focus:border-blue-500 flex-1"
            >
              <option v-for="opt in craftOptions" :key="opt.value" :value="opt.value">
                {{ opt.label }}
              </option>
            </select>
          </div>
        </div>
      </div>
    </aside>

    <!-- 右侧：任务列表 -->
    <div class="flex-1 min-w-0 flex flex-col gap-5">
      <BountyCard
        v-for="task in tasks"
        :key="task.id"
        :title="task.title"
        :publishTime="task.publishTime"
        :tags="task.tags"
        :description="task.description"
        :deadline="task.deadline"
      />
    </div>
  </div>
</template>