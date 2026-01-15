<script setup>
defineProps({
  bounty: {
    type: Object,
    required: true
  }
})

const statusMap = {
  open: { text: '待接单', class: 'bg-amber-50 text-amber-600' },
  in_progress: { text: '进行中', class: 'bg-blue-50 text-blue-600' },
  completed: { text: '已完成', class: 'bg-green-50 text-green-600' },
  closed: { text: '已关闭', class: 'bg-gray-100 text-gray-500' }
}

function getStatusInfo(status) {
  return statusMap[status] || statusMap.open
}

function formatReward(reward) {
  return new Intl.NumberFormat('zh-CN', {
    style: 'currency',
    currency: 'CNY',
    minimumFractionDigits: 0
  }).format(reward)
}

function formatDate(dateStr) {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  return date.toLocaleDateString('zh-CN')
}
</script>

<template>
  <div
    class="bg-white rounded-xl p-6 shadow-sm hover:shadow-md hover:-translate-y-1 transition-all duration-300 cursor-pointer"
  >
    <!-- 头部：标题 + 金额 -->
    <div class="flex justify-between items-start mb-4">
      <div class="flex-1 mr-4">
        <h3 class="text-lg font-semibold text-slate-800 mb-2 line-clamp-2">
          {{ bounty.title }}
        </h3>
        <div class="flex items-center gap-4 text-sm text-slate-500">
          <span class="flex items-center gap-1">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
            </svg>
            {{ bounty.created_by || '匿名' }}
          </span>
          <span class="flex items-center gap-1">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            {{ formatDate(bounty.created_at) }}
          </span>
        </div>
      </div>
      <div class="bg-amber-50 text-amber-600 px-4 py-2 rounded-lg font-bold text-lg whitespace-nowrap">
        {{ formatReward(bounty.reward) }}
      </div>
    </div>

    <!-- 描述 -->
    <p class="text-slate-600 text-sm leading-relaxed mb-4 line-clamp-2">
      {{ bounty.description || '暂无描述' }}
    </p>

    <!-- 底部：状态 -->
    <div class="flex justify-between items-center pt-4 border-t border-slate-100">
      <span
        :class="[
          'px-3 py-1 rounded-full text-xs font-medium',
          getStatusInfo(bounty.status).class
        ]"
      >
        {{ getStatusInfo(bounty.status).text }}
      </span>
      <span class="text-xs text-slate-400">
        ID: {{ bounty.id }}
      </span>
    </div>
  </div>
</template>
