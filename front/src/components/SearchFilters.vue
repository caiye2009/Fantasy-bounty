<script setup>
import { computed } from 'vue'

const props = defineProps({
  filters: {
    type: Array,
    default: () => []
  },
  selected: {
    type: Object,
    default: () => ({})
  }
})

const emit = defineEmits(['change'])

// 检查某个值是否被选中
const isSelected = (field, value) => {
  const current = props.selected[field]
  if (!current) return false
  if (Array.isArray(current)) {
    return current.includes(value)
  }
  return current === value
}

// 处理 terms 类型筛选器点击
const handleTermsSelect = (field, value) => {
  const current = props.selected[field] || []
  let newValue

  if (Array.isArray(current)) {
    if (current.includes(value)) {
      // 取消选择
      newValue = current.filter(v => v !== value)
    } else {
      // 添加选择
      newValue = [...current, value]
    }
  } else {
    newValue = [value]
  }

  emit('change', field, newValue.length > 0 ? newValue : null)
}

// 处理 range 类型筛选器点击（单选）
const handleRangeSelect = (field, value) => {
  const current = props.selected[field]
  if (current === value) {
    // 取消选择
    emit('change', field, null)
  } else {
    // 选择新值
    emit('change', field, value)
  }
}

// 有数据的筛选项
const visibleFilters = computed(() => {
  return props.filters.filter(f => f.buckets && f.buckets.length > 0)
})
</script>

<template>
  <div class="space-y-4">
    <div
      v-for="filter in visibleFilters"
      :key="filter.field"
      class="bg-white rounded-lg p-4 shadow-sm"
    >
      <h4 class="font-medium text-gray-700 mb-3 text-sm">{{ filter.label }}</h4>

      <!-- Terms 类型筛选器（多选） -->
      <div v-if="filter.type === 'terms'" class="space-y-1">
        <label
          v-for="bucket in filter.buckets"
          :key="bucket.key"
          class="flex items-center gap-2 cursor-pointer hover:bg-gray-50 p-1.5 rounded text-sm"
        >
          <input
            type="checkbox"
            :checked="isSelected(filter.field, bucket.key)"
            @change="handleTermsSelect(filter.field, bucket.key)"
            class="rounded text-blue-500 focus:ring-blue-500"
          />
          <span class="flex-1 text-gray-600">{{ bucket.label }}</span>
          <span class="text-xs text-gray-400">({{ bucket.docCount }})</span>
        </label>
      </div>

      <!-- Range 类型筛选器（单选） -->
      <div v-else-if="filter.type === 'range'" class="space-y-1">
        <button
          v-for="bucket in filter.buckets"
          :key="bucket.key"
          @click="handleRangeSelect(filter.field, bucket.key)"
          :class="[
            'w-full text-left px-3 py-1.5 rounded text-sm transition-colors',
            isSelected(filter.field, bucket.key)
              ? 'bg-blue-50 text-blue-600 border border-blue-200'
              : 'hover:bg-gray-50 border border-transparent text-gray-600'
          ]"
        >
          <span class="flex items-center justify-between">
            <span>{{ bucket.label }}</span>
            <span class="text-xs text-gray-400">({{ bucket.docCount }})</span>
          </span>
        </button>
      </div>
    </div>

    <!-- 空状态 -->
    <div
      v-if="visibleFilters.length === 0"
      class="bg-white rounded-lg p-4 shadow-sm text-center text-gray-400 text-sm"
    >
      暂无可用筛选项
    </div>
  </div>
</template>
