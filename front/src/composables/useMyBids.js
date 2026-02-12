import { ref, watch } from 'vue'
import { fetchMyBids } from '@/api/bid'
import { formatDateTime, formatDate } from '@/utils/format'

// 状态选项
export const statusOptions = [
  { label: '全部', value: '' },
  { label: '审核中', value: 'pending' },
  { label: '进行中', value: 'in_progress' },
  { label: '待验收', value: 'pending_acceptance' },
  { label: '已完成', value: 'completed' }
]

// 转换投标数据
const transformBidItem = (item) => ({
  id: item.id,
  bountyId: item.bountyId,
  bountyTitle: item.bountyProductName || '未知悬赏',
  bountyProductCode: item.bountyProductCode || '',
  bountyType: item.bountyType,
  bidTime: formatDateTime(item.createdAt),
  status: item.status,
  bidAmount: item.bidPrice,
  deadline: formatDate(item.bidDeadline),
  wovenSpec: item.wovenSpec,
  knittedSpec: item.knittedSpec
})

export function useMyBids() {
  const selectedStatus = ref('')
  const myBids = ref([])
  const loading = ref(false)
  const error = ref(null)

  const loadMyBids = async () => {
    loading.value = true
    error.value = null
    myBids.value = []
    try {
      const result = await fetchMyBids(selectedStatus.value, 1, 50)
      myBids.value = result.data.map(transformBidItem)
    } catch (e) {
      console.error('加载投标列表失败:', e)
      error.value = e.message
      myBids.value = []
    } finally {
      loading.value = false
    }
  }

  // 监听筛选条件变化
  watch(selectedStatus, () => {
    loadMyBids()
  })

  return {
    selectedStatus,
    myBids,
    loading,
    error,
    loadMyBids,
    statusOptions,
  }
}
