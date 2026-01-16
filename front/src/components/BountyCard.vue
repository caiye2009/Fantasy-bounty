<script setup>
import { CalendarDays } from 'lucide-vue-next'

const props = defineProps({
  title: { type: String, required: true },
  publishTime: { type: String, required: true },
  tags: { type: Array, default: () => [] },
  description: { type: String, required: true },
  deadline: { type: String, required: true }
})
</script>

<template>
  <div
    class="bg-white rounded-2xl shadow-md hover:-translate-y-2 transform transition-transform duration-300
           p-8 w-full flex flex-col justify-between cursor-pointer gap-4"
  >
    <!-- 标题 + 发布时间 -->
    <div class="space-y-3">
      <h3 class="text-xl font-bold text-gray-900">{{ title }}</h3>
      <p class="text-gray-500 text-sm">{{ publishTime }}</p>

      <!-- Tag 行 -->
      <div class="flex flex-wrap gap-2">
        <span
          v-for="(tag, index) in tags"
          :key="index"
          class="bg-gray-200 text-gray-700 text-xs px-2 py-1 rounded"
        >
          {{ tag }}
        </span>
      </div>

      <!-- 描述 -->
      <p class="text-gray-800 text-sm line-clamp-3">
        {{ description }}
      </p>
    </div>

    <!-- 截止日期 -->
    <p class="text-gray-500 text-sm mt-auto flex items-center gap-1">
      <CalendarDays :size="16" />
      截止日期：{{ deadline }}
    </p>
  </div>
</template>