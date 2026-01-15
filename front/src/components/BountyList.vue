<script setup>
import BountyCard from './BountyCard.vue'

defineProps({
  bounties: {
    type: Array,
    required: true
  },
  loading: {
    type: Boolean,
    default: false
  },
  error: {
    type: String,
    default: null
  }
})

defineEmits(['retry'])
</script>

<template>
  <!-- 加载状态 -->
  <div v-if="loading" class="flex flex-col items-center justify-center py-20">
    <div class="w-10 h-10 border-4 border-blue-500 border-t-transparent rounded-full animate-spin mb-4"></div>
    <p class="text-slate-500">加载中...</p>
  </div>

  <!-- 错误状态 -->
  <div v-else-if="error" class="flex flex-col items-center justify-center py-20">
    <svg class="w-16 h-16 text-red-400 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
    </svg>
    <p class="text-slate-600 mb-4">{{ error }}</p>
    <button
      @click="$emit('retry')"
      class="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors"
    >
      重试
    </button>
  </div>

  <!-- 空状态 -->
  <div v-else-if="bounties.length === 0" class="flex flex-col items-center justify-center py-20">
    <svg class="w-16 h-16 text-slate-300 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
    </svg>
    <p class="text-slate-500 text-lg mb-2">暂无悬赏任务</p>
    <p class="text-slate-400 text-sm">稍后再来看看吧</p>
  </div>

  <!-- 列表 -->
  <div v-else class="flex flex-col gap-5">
    <BountyCard
      v-for="bounty in bounties"
      :key="bounty.id"
      :bounty="bounty"
    />
  </div>
</template>
